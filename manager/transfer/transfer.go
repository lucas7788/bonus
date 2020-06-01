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
	verifyTxQueue      chan verifyParam
	hasTransferedOntid map[string]bool
	CloseChan          chan bool
	rwLock             *sync.RWMutex
	TransferStatus     common.TransferStatus
	TokenType          string
	db                 *bonus_db.BonusDB
	stopChan           chan bool
}

type verifyParam struct {
	TxHash   string
	TxHex    []byte
	needSend bool
}

func NewTxHandleTask(tokenType string, db *bonus_db.BonusDB, txQueueSize int, stopChan chan bool) *TxHandleTask {
	verifyQueue := make(chan verifyParam, txQueueSize/2)
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
	defer close(self.verifyTxQueue)
	collectDataLen := len(collectData)
	txCaches, err := self.db.QueryTxInfo()
	if err != nil {
		log.Errorf("[StartTransfer] QueryTxInfo failed: %s", err)
		return err
	}
	newTxHexMap := make(map[string][]byte)
	limit := 500
	txInfoArr := make([]*common.TransactionInfo, limit)
	txInfoArrIndex := 0
	var txHash string
	var txHex []byte
	index := 0
	for addr, amt := range collectData {
		if txCaches[addr] != nil && txCaches[addr].TxStatus == common.TxSuccess {
			continue
		}
		var err error
		if txCaches[addr] != nil && txCaches[addr].TxStatus != common.TxSuccess {
			txHex, err := hex.DecodeString(txCaches[addr].TxHex)
			if err != nil {
				panic(err)
			}
			self.verifyTxQueue <- verifyParam{
				TxHash:   txHash,
				TxHex:    txHex,
				needSend: true,
			}
		} else if txCaches[addr] == nil {
			txHash, txHex, err = mana.NewWithdrawTx(addr, amt, excel.TokenType)
			if err != nil {
				log.Errorf("[StartTransfer] NewWithdrawTx failed: %s", err)
				return err
			}
			newTxHexMap[addr] = txHex

			txInfoArr[txInfoArrIndex] = &common.TransactionInfo{
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
			if txInfoArrIndex == limit-1 || index == collectDataLen-1 {
				err := self.db.InsertTxInfoSql(txInfoArr)
				if err != nil {
					log.Errorf("[StartTransfer] InsertTxInfoSql failed: %s", err)
					return err
				}
				for j := 0; j < len(txInfoArr); j++ {
					if txInfoArr[j] == nil {
						continue
					}
					txHash, err := mana.SendTx(newTxHexMap[txInfoArr[j].Address])
					if err != nil {
						log.Errorf("[StartTransfer] SendTx error: %s", err)
						return err
					}
					log.Infof("tx send success, txhash:%s", txHash)
					self.verifyTxQueue <- verifyParam{
						TxHash: txHash,
					}
					txInfoArr[j] = nil
				}
				txInfoArrIndex = 0
			}
		}
		index += 1
		select {
		case <-self.stopChan:
			log.Infof("[StartTransfer] stopChan, address: %s", addr)
			return nil
		default:
			continue
		}
	}
	return nil
}

func (this *TxHandleTask) UpdateTxInfoTable(mana interfaces.WithdrawManager, eatp *common.ExcelParam) (map[int]bool, error) {
	txInfos := make([]*common.TransactionInfo, 0)
	hasBuildTxId := make(map[int]bool)
	//update tx info table
	for _, trParam := range eatp.BillList {
		if trParam == nil {
			log.Errorf("trparam is nil")
			continue
		}
		tx, err := this.db.QueryTxHexByExcelAndAddr(eatp.EventType, trParam.Address, trParam.Id)
		if err != nil {
			log.Errorf("QueryTxHexByExcelAndAddr error: %s, eventType:%s, address:%s, id: %d", err, eatp.EventType, trParam.Address, trParam.Id)
			continue
		}
		if tx == nil {
			txInfo := &common.TransactionInfo{
				Id:              trParam.Id,
				EventType:       eatp.EventType,
				TokenType:       eatp.TokenType,
				ContractAddress: eatp.ContractAddress,
				Address:         trParam.Address,
				Amount:          trParam.Amount,
				NetType:         mana.GetNetType(),
				TxResult:        common.NotBuild,
			}
			txInfos = append(txInfos, txInfo)
		} else {
			hasBuildTxId[trParam.Id] = true
		}
	}
	if len(txInfos) > 0 {
		err := this.db.InsertTxInfoSql(txInfos)
		if err != nil {
			log.Errorf("InsertTxInfoSql error:%s", err)
			return nil, err
		}
	}
	return hasBuildTxId, nil
}

func (self *TxHandleTask) WaitClose() {
	<-self.CloseChan
}

func (self *TxHandleTask) StartVerifyTxTask(mana interfaces.WithdrawManager) {

	for {
		select {
		case verifyP, ok := <-self.verifyTxQueue:
			if !ok {
				self.TransferStatus = common.Transfered
				log.Info("exit StartVerifyTxTask gorountine")
				return
			}
			var boo bool
			var err error
			boo, err = mana.VerifyTx(verifyP.TxHash, config.RetryLimit)
			if err != nil || !boo {
				if verifyP.needSend {
					_, err = mana.SendTx(verifyP.TxHex)
					if err != nil {
						log.Errorf("[StartVerifyTxTask] send tx failed: %s", err)
						continue
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
	}
}
