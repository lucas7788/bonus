package transfer

import (
	"strings"
	"sync"
	"time"

	"github.com/ontio/bonus/bonus_db"
	"github.com/ontio/bonus/common"
	"github.com/ontio/bonus/config"
	"github.com/ontio/bonus/manager/interfaces"
	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
)

type TxHandleTask struct {
	TransferQueue      chan *common.TransferParam
	verifyTxQueue      chan *VerifyParam
	hasTransferedOntid map[string]bool
	CloseChan          chan bool
	waitVerify         chan bool
	rwLock             *sync.RWMutex
	TransferStatus     common.TransferStatus
	TokenType          string
	db                 *bonus_db.BonusDB
	StopTransferChan   chan bool
	stopVerifyChan     chan bool
}

type VerifyParam struct {
	Id        int
	TxHash    string
	EventType string
	Address   string
}

func NewTxHandleTask(tokenType string, db *bonus_db.BonusDB, txQueueSize int) *TxHandleTask {
	transferQueue := make(chan *common.TransferParam, txQueueSize)
	verifyQueue := make(chan *VerifyParam, txQueueSize/2)
	return &TxHandleTask{
		TransferQueue:    transferQueue,
		verifyTxQueue:    verifyQueue,
		TransferStatus:   common.Transfering,
		CloseChan:        make(chan bool),
		waitVerify:       make(chan bool),
		StopTransferChan: make(chan bool),
		stopVerifyChan:   make(chan bool),
		TokenType:        tokenType,
		db:               db,
	}
}

func (this *TxHandleTask) UpdateTxInfoTable(mana interfaces.WithdrawManager, eatp *common.ExcelParam) (map[int]bool, error) {
	txInfos := make([]*common.TransactionInfo, 0)
	hasBuildTxId := make(map[int]bool)
	//update tx info table
	for _, trParam := range eatp.BillList {
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

func (self *TxHandleTask) StartHandleTransferTask(mana interfaces.WithdrawManager, eventType string) {
	for {
		select {
		case <-self.StopTransferChan:
			self.stopVerifyChan <- true
			log.Info("[StartHandleTransferTask] exit transfer")
			return
		case param, ok := <-self.TransferQueue:
			if !ok || param == nil {
				self.exit()
				return
			}
			var txHex []byte
			var err error
			txInfo, err := self.db.QueryTxHexByExcelAndAddr(eventType, param.Address, param.Id)
			if err != nil {
				log.Errorf("QueryTxHexByOntid failed,address: %s, error: %s", param.Address, err)
				continue
			}

			if txInfo != nil && txInfo.TxResult == common.TxSuccess {
				continue
			}
			//if tx verify failed, here should be verify again
			if (txInfo != nil && txInfo.TxResult == common.TxFailed && txInfo.TxHash != "") ||
				(txInfo != nil && txInfo.TxResult == common.OneTransfering && txInfo.TxHash != "") ||
				(txInfo != nil && txInfo.TxResult == common.SendFailed && txInfo.TxHash != "") {

				err = self.db.UpdateTxResult(eventType, param.Address, param.Id, common.OneTransfering, 0, "")
				if err != nil {
					log.Errorf("UpdateTxResult error: %s, eventType: %s, address: %s, projectId: %d, txHash: %s, txresult: %d",
						err, eventType, param.Address, txInfo.TxHash, byte(common.OneTransfering))
				}

				boo, err := mana.VerifyTx(txInfo.TxHash, 1)
				if boo {
					log.Infof("Failed transactions revalidate success, txhash: %s", txInfo.TxHash)
					ti, err := mana.GetTxTime(txInfo.TxHash)
					if err != nil {
						log.Errorf("GetTxTime error: %s", err)
						continue
					}
					err = self.db.UpdateTxResult(eventType, param.Address, param.Id, common.TxSuccess, ti, "")
					if err != nil {
						log.Errorf("UpdateTxResult failed, txhash: %s, error: %s", txInfo.TxHash, err)
					}
					continue
				}
				log.Infof("Failed transactions revalidate failed, txhash: %s, error: %s", txInfo.TxHash, err)
			}
			if txInfo != nil && txInfo.TxHex != "" {
				txHex, err = common2.HexToBytes(txInfo.TxHex)
				if err != nil {
					log.Errorf("QueryTxHexByOntid HexToBytes failed, error: %s", err)
					continue
				}
			}
			//build tx
			if txHex == nil {
				var txHash string
				txHash, txHex, err = mana.NewWithdrawTx(param.Address, param.Amount, "")
				if err != nil && strings.Contains(err.Error(), config.InSufficientBalance) {
					self.exit()
					log.Errorf("Address: %s, amount:%s, error: %s", param.Address, param.Amount, err)
					return
				}
				if err != nil || txHash == "" || txHex == nil {
					log.Errorf("Build Transfer Tx failed,address: %s,txHash: %s, err: %s", param.Address, txHash, err)
					err := self.db.UpdateTxResult(eventType, param.Address, param.Id, common.BuildTxFailed, 0, err.Error())
					if err != nil {
						log.Errorf("UpdateTxResult error: %s, eventType: %s, address: %s", err, eventType, param.Address)
					}
					continue
				}
				log.Debugf("tx build success, txhash: %s", txHash)
				err := self.db.UpdateTxInfo(txHash, common2.ToHexString(txHex), common.OneTransfering, eventType, param.Address, param.Id)
				if err != nil {
					log.Errorf("UpdateTxInfo error: %s, event type:%s, address: %s", err, eventType, param.Address)
					continue
				}
				log.Infof("txHash:%s, txHex:%s, address:%s, amount:%s", txHash, common2.ToHexString(txHex), param.Address, param.Amount)
			}
			//send tx
			retry := 0
			var hash string
			for {
				hash, err = mana.SendTx(txHex)
				if err != nil && retry < config.RetryLimit {
					if err != nil {
						log.Errorf("SendTx error :%s, retry:%d", err, retry)
						if strings.Contains(err.Error(), "insufficient funds for gas * price + value") ||
							strings.Contains(err.Error(), "duplicated transaction detected") {
							mana.Stop()
							return
						}
					}
					retry += 1
					time.Sleep(time.Duration(retry*config.SleepTime) * time.Second)
					continue
				} else {
					break
				}
			}
			if err != nil || hash == "" {
				log.Errorf("SendTx error: %s, txhash: %s", err, hash)
				//save txHex
				err = self.db.UpdateTxResult(eventType, param.Address, param.Id, common.SendFailed, 00, err.Error())
				if err != nil {
					log.Errorf("UpdateTxResult error: %s, eventType: %s, address: %s, txHash: %s, txresult: %d",
						err, eventType, param.Address, txInfo.TxHash, byte(common.SendFailed))
				}
				continue
			} else {
				//save txHex
				err = self.db.UpdateTxResult(eventType, param.Address, param.Id, common.OneTransfering, 0, "")
				if err != nil {
					log.Errorf("UpdateTxResult error: %s, eventType: %s, address: %s, projectId: %d, txHash: %s, txresult: %d",
						err, eventType, param.Address, txInfo.TxHash, byte(common.OneTransfering))
				}
			}
			log.Infof("tx send success, txhash: %s, time: %d", hash, time.Now().Unix())
			self.verifyTxQueue <- &VerifyParam{
				Id:        param.Id,
				TxHash:    hash,
				Address:   param.Address,
				EventType: eventType,
			}
		}
	}
}

func (self *TxHandleTask) StartVerifyTxTask(mana interfaces.WithdrawManager) {

	for {
		select {
		case <-self.stopVerifyChan:
			self.TransferStatus = common.Transfered
			log.Info("[StartVerifyTxTask] exit verify")
			return
		case verifyParam, ok := <-self.verifyTxQueue:
			if !ok || verifyParam.TxHash == "" {
				self.TransferStatus = common.Transfered
				self.waitVerify <- true
				log.Info("exit StartVerifyTxTask gorountine")
				return
			}
			boo, err := mana.VerifyTx(verifyParam.TxHash, config.RetryLimit)
			if !boo {
				//save failed tx to bonus_db
				err := self.db.UpdateTxResult(verifyParam.EventType, verifyParam.Address, verifyParam.Id, common.TxFailed, 0, err.Error())
				if err != nil {
					log.Errorf("UpdateTxResult error: %s, txHash: %s", err, verifyParam.TxHash)
				}
				log.Errorf("VerifyTx failed, txhash: %s, error: %s", verifyParam.TxHash, err)
				continue
			}
			ti, err := mana.GetTxTime(verifyParam.TxHash)
			if err != nil {
				log.Errorf("GetTxTime error: %s", err)
				continue
			}
			//update bonus_db
			err = self.db.UpdateTxResult(verifyParam.EventType, verifyParam.Address, verifyParam.Id, common.TxSuccess, ti, "success")
			if err != nil {
				log.Errorf("UpdateTxResult error: %s, txHash: %s", err, verifyParam.TxHash)
			}
			log.Debugf("verify tx success, txhash: %s", verifyParam.TxHash)
		}
	}
}
