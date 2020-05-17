package transfer

import (
	"github.com/ontio/bonus/bonus_db"
	"github.com/ontio/bonus/common"
	"github.com/ontio/bonus/config"
	"github.com/ontio/bonus/manager/interfaces"
	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"math/big"
	"sync"
)

type TxHandleTask struct {
	verifyTxQueue      chan string
	hasTransferedOntid map[string]bool
	CloseChan          chan bool
	waitVerify         chan bool
	rwLock             *sync.RWMutex
	TransferStatus     common.TransferStatus
	TokenType          string
	db                 *bonus_db.BonusDB
	stopChan           chan bool
	stopVerifyChan     chan bool
}

type VerifyParam struct {
	Id        int
	TxHash    string
	EventType string
	Address   string
}

func NewTxHandleTask(tokenType string, db *bonus_db.BonusDB, txQueueSize int, stopChan chan bool) *TxHandleTask {
	verifyQueue := make(chan string, txQueueSize/2)
	return &TxHandleTask{
		verifyTxQueue:  verifyQueue,
		TransferStatus: common.Transfering,
		CloseChan:      make(chan bool),
		waitVerify:     make(chan bool),
		stopVerifyChan: make(chan bool),
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

func (self *TxHandleTask) StartTxTask(mana interfaces.WithdrawManager, txCaches map[string]*common.TxCache, excel *common.ExcelParam, collectData map[string]*big.Int) (map[string]*common.TxCache, error) {
	defer close(self.verifyTxQueue)
	collectDataLen := len(collectData)
	if txCaches == nil {
		txCache := make(map[string]*common.TxCache, collectDataLen)
		txCache, err := self.db.QueryTxInfo(txCache)
		if err != nil {
			log.Errorf("[StartTransfer] QueryTxInfo failed: %s", err)
			return nil, err
		}
		txCaches = txCache
	} else {
		total, err := self.db.QueryTxInfoNum()
		if err != nil {
			log.Errorf("[StartTransfer] QueryTxInfoNum failed: %s", err)
			return nil, err
		}
		txCaches, err = self.updateTxCache(total, txCaches)
		if err != nil {
			return nil, err
		}
	}
	txInfoArr := make([]*common.TransactionInfo, 30)
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
			txHash, err = mana.SendTx(txCaches[addr].TxHex)
			if err != nil {
				log.Errorf("[StartTransfer] SendTx error: %s", err)
				return nil, err
			}
		} else if txCaches[addr] == nil {
			txHash, txHex, err = mana.NewWithdrawTx(addr, amt, excel.TokenType)
			if err != nil {
				log.Errorf("[StartTransfer] NewWithdrawTx failed: %s", err)
				return nil, err
			}
			txCaches[addr] = &common.TxCache{
				Addr:     addr,
				TxHash:   txHash,
				TxHex:    txHex,
				TxStatus: common.OneTransfering,
			}
			txInfoArr[txInfoArrIndex] = &common.TransactionInfo{
				Id:              index,
				NetType:         excel.NetType,
				EventType:       excel.EventType,
				TokenType:       excel.TokenType,
				ContractAddress: excel.ContractAddress,
				Address:         addr,
				Amount:          addr,
				TxHash:          txHash,
				TxTime:          0,
				TxHex:           common2.ToHexString(txHex),
				TxResult:        common.OneTransfering,
				ErrorDetail:     "",
			}
			txInfoArrIndex += 1
			if txInfoArrIndex == 29 || index == collectDataLen-1 {
				err := self.db.InsertTxInfoSql(txInfoArr)
				if err != nil {
					log.Errorf("[StartTransfer] InsertTxInfoSql failed: %s", err)
					return nil, err
				}
				for j := 0; j < len(txInfoArr); j++ {
					if txInfoArr[j] == nil {
						continue
					}
					txHash, err := mana.SendTx(txCaches[txInfoArr[j].Address].TxHex)
					if err != nil {
						log.Errorf("[StartTransfer] SendTx error: %s", err)
						return nil, err
					}
					log.Infof("tx send success, txhash:%s", txHash)
					self.verifyTxQueue <- txHash
					txInfoArr[j] = nil
				}
				txInfoArrIndex = 0
			}
		}
		index += 1
		select {
		case <-self.stopChan:
			log.Infof("[StartTransfer] stopChan, address: %d", addr)
			return txCaches, nil
		default:
			continue
		}
	}
	return txCaches, nil
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

func (self *TxHandleTask) exit() {
	close(self.verifyTxQueue)
	log.Infof("1. close(self.verifyTxQueue)")
	self.CloseChan <- true
	<-self.waitVerify
	log.Info("exit StartHandleTransferTask gorountine")
}

func (self *TxHandleTask) StartVerifyTxTask(mana interfaces.WithdrawManager) {

	for {
		select {
		case <-self.stopVerifyChan:
			self.TransferStatus = common.Transfered
			log.Info("[StartVerifyTxTask] exit verify")
			return
		case txHash, ok := <-self.verifyTxQueue:
			if !ok || txHash == "" {
				self.TransferStatus = common.Transfered
				self.waitVerify <- true
				log.Info("exit StartVerifyTxTask gorountine")
				return
			}
			boo, err := mana.VerifyTx(txHash, config.RetryLimit)
			if !boo {
				//save failed tx to bonus_db
				err := self.db.UpdateTxResultByTxHash(txHash, common.TxFailed, 0, err.Error())
				if err != nil {
					log.Errorf("UpdateTxResult error: %s, txHash: %s", err, txHash)
				}
				log.Errorf("VerifyTx failed, txhash: %s, error: %s", txHash, err)
				continue
			}
			ti, err := mana.GetTxTime(txHash)
			if err != nil {
				log.Errorf("GetTxTime error: %s", err)
				continue
			}
			//update bonus_db
			err = self.db.UpdateTxResultByTxHash(txHash, common.TxSuccess, ti, "success")
			if err != nil {
				log.Errorf("UpdateTxResult error: %s, txHash: %s", err, txHash)
				continue
			}
			log.Debugf("verify tx success, txhash: %s", txHash)
		}
	}
}
