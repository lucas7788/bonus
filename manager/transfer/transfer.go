package transfer

import (
	"encoding/hex"
	"github.com/ontio/bonus/bonus_db"
	"github.com/ontio/bonus/common"
	"github.com/ontio/bonus/config"
	"github.com/ontio/bonus/manager/interfaces"
	"github.com/ontio/bonus/utils"
	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"math/big"
	"sync"
)

type TxHandleTask struct {
	verifyTxQueue  chan verifyParam
	CloseChan      chan bool
	rwLock         *sync.RWMutex
	TransferStatus common.TransferStatus
	TokenType      string
	db             *bonus_db.BonusDB
	stopChan       chan bool
}

type verifyParam struct {
	TxHash   string
	TxHex    []byte
	needSend bool
}

func NewTxHandleTask(tokenType string, db *bonus_db.BonusDB, txQueueSize int, stopChan chan bool) *TxHandleTask {
	verifyQueue := make(chan verifyParam, txQueueSize)
	return &TxHandleTask{
		verifyTxQueue:  verifyQueue,
		TransferStatus: common.Transfering,
		CloseChan:      make(chan bool),
		TokenType:      tokenType,
		db:             db,
		stopChan:       stopChan,
	}
}

func (self *TxHandleTask) updateTxCache(total int, txCaches map[string]*common.TxCache) (map[string]*common.TxCache, error) {
	for addr, val := range txCaches {
		if val.TxStatus != common.TxSuccess {
			txResult, err := self.db.QueryTxResult(val.Addr)
			if err != nil {
				return nil, err
			}
			txCaches[addr].TxStatus = txResult
		}
	}
	return txCaches, nil
}

func (self *TxHandleTask) StartTxTask(mana interfaces.WithdrawManager, excel *common.ExcelParam, collectData map[string]*big.Int, decimal uint64) error {
	collectDataLen := len(collectData)
	txCaches, err := self.db.QueryTxInfo()
	if err != nil {
		log.Errorf("[StartTransfer] QueryTxInfo failed: %s", err)
		return err
	}
	newTxHexMap := make(map[string][]byte)
	limit := 200
	if excel.TokenType == config.ETH || excel.TokenType == config.ERC20 {
		limit = 5
	}
	txInfoArr := make([]common.TransactionInfo, limit)
	txInfoArrIndex := 0
	index := 0
	log.Infof("collectData length: %d, txCaches length: %d", len(collectData), len(txCaches))
	for addr, amt := range collectData {
		index += 1
		if txCaches[addr] != nil && txCaches[addr].TxStatus == common.TxSuccess {
			continue
		}
		if txCaches[addr] != nil && txCaches[addr].TxStatus != common.TxSuccess {
			txHex, err := hex.DecodeString(txCaches[addr].TxHex)
			if err != nil {
				panic(err)
			}
			self.verifyTxQueue <- verifyParam{
				TxHash:   txCaches[addr].TxHash,
				TxHex:    txHex,
				needSend: true,
			}
		} else if txCaches[addr] == nil {
			txHash, txHex, err := mana.NewWithdrawTx(addr, amt, excel.TokenType)
			if err != nil {
				log.Errorf("[StartTransfer] NewWithdrawTx failed: %s", err)
				return err
			}
			newTxHexMap[addr] = txHex
			txInfoArr[txInfoArrIndex] = common.TransactionInfo{
				NetType:         excel.NetType,
				EventType:       excel.EventType,
				TokenType:       excel.TokenType,
				ContractAddress: excel.ContractAddress,
				Address:         addr,
				Amount:          utils.ToStringByPrecise(amt, decimal),
				TxHash:          txHash,
				TxTime:          0,
				TxHex:           common2.ToHexString(txHex),
				TxResult:        common.OneTransfering,
				ErrorDetail:     "",
			}
			txInfoArrIndex += 1
			if txInfoArrIndex >= limit-1 || index >= collectDataLen-1 {
				err = self.insertAndSendTx(mana, txInfoArr, newTxHexMap)
				if err != nil {
					log.Errorf("[StartTransfer] insertAndSendTx failed: %s", err)
					return err
				}
				txInfoArr = make([]common.TransactionInfo, limit)
				txInfoArrIndex = 0
			}
		}
		select {
		case <-self.stopChan:
			log.Infof("[StartTransfer] stopChan, address: %s", addr)
			self.TransferStatus = common.Stop
			close(self.verifyTxQueue)
			return nil
		default:
			continue
		}
	}
	err = self.insertAndSendTx(mana, txInfoArr, newTxHexMap)
	if err != nil {
		log.Infof("[StartTransfer] for end insertAndSendTx, error: %s", err)
		return err
	}
	close(self.verifyTxQueue)
	return nil
}

func (this *TxHandleTask) insertAndSendTx(mana interfaces.WithdrawManager, txInfoArr []common.TransactionInfo, newTxHexMap map[string][]byte) error {
	err := this.db.InsertTxInfoSql(txInfoArr)
	if err != nil {
		log.Errorf("[StartTransfer] InsertTxInfoSql failed: %s", err)
		return err
	}
	for i := 0; i < len(txInfoArr); i++ {
		if txInfoArr[i].TxHash == "" {
			continue
		}
		txHash, err := mana.SendTx(newTxHexMap[txInfoArr[i].Address])
		if err != nil {
			log.Errorf("SendTx failed: %s, txhash: %s\n", err, txHash)
			return err
		}
		this.verifyTxQueue <- verifyParam{
			TxHash: txHash,
		}
	}
	return nil
}

func (self *TxHandleTask) WaitClose() {
	<-self.CloseChan
}

func (self *TxHandleTask) StartVerifyTxTask(mana interfaces.WithdrawManager) {
	for verifyP := range self.verifyTxQueue {
		var boo bool
		var err error
		boo, err = mana.VerifyTx(verifyP.TxHash, config.RetryLimit)
		if err != nil || !boo {
			if verifyP.needSend {
				_, err = mana.SendTx(verifyP.TxHex)
				if err != nil {
					log.Errorf("[StartVerifyTxTask] send tx failed: %s", err)
				}
				boo, err = mana.VerifyTx(verifyP.TxHash, config.RetryLimit)
				if err != nil {
					log.Errorf("[StartVerifyTxTask] VerifyTx failed: %s", err)
					continue
				}
			}
		}
		if !boo {
			//save failed tx to bonus_db
			err := self.db.UpdateTxResultByTxHash(verifyP.TxHash, common.TxFailed, 0, err.Error())
			if err != nil {
				log.Errorf("UpdateTxResult error: %s, txHash: %s", err, verifyP.TxHash)
			}
			log.Errorf("VerifyTx failed, txhash: %s, error: %s", verifyP.TxHash, err)
			continue
		}
		ti, err := mana.GetTxTime(verifyP.TxHash)
		if err != nil {
			log.Errorf("GetTxTime error: %s", err)
			continue
		}
		//update bonus_db
		err = self.db.UpdateTxResultByTxHash(verifyP.TxHash, common.TxSuccess, ti, "success")
		if err != nil {
			log.Errorf("UpdateTxResult error: %s, txHash: %s", err, verifyP.TxHash)
			continue
		}
		log.Debugf("verify tx success, txhash: %s", verifyP.TxHash)
	}
	log.Info("exit StartVerifyTxTask gorountine")
	self.TransferStatus = common.Transfered
}
