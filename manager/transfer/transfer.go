package transfer

import (
	"github.com/ontio/bonus/bonus_db"
	"github.com/ontio/bonus/common"
	"github.com/ontio/bonus/config"
	"github.com/ontio/bonus/manager/interfaces"
	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"sync"
	"time"
)

type TxHandleTask struct {
	TransferQueue      chan *common.TransferParam
	verifyTxQueue      chan string
	hasTransferedOntid map[string]bool
	closeChan          chan bool
	rwLock             *sync.RWMutex
}

func NewTxHandleTask() *TxHandleTask {
	transferQueue := make(chan *common.TransferParam, config.TRANSFER_QUEUE_SIZE)
	verifyQueue := make(chan string, config.VERIFY_TX_QUEUE_SIZE)
	return &TxHandleTask{
		TransferQueue: transferQueue,
		verifyTxQueue: verifyQueue,
	}
}

func (self *TxHandleTask) WaitClose() {
	<-self.closeChan
}

func (self *TxHandleTask) StartHandleTransferTask(mana interfaces.WithdrawManager, eventType string) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("recover info: ", r)
		}
	}()
	for {
		select {
		case param, ok := <-self.TransferQueue:
			if !ok || param == nil {
				close(self.verifyTxQueue)
				log.Infof("close(self.verifyTxQueue)")
				return
			}
			var txHex []byte
			var err error
			txInfo, err := bonus_db.QueryTxHexByExcelAndAddr(eventType, param.Address)
			if err != nil {
				log.Errorf("QueryTxHexByOntid failed,address: %s, error: %s", param.Address, err)
				continue
			}
			if txInfo != nil && txInfo.TxResult == common.TxSuccess {
				continue
			}
			//if tx verify failed, here should be verify again
			if (txInfo != nil && txInfo.TxResult == common.TxFailed && txInfo.TxHash != "") ||
				(txInfo != nil && txInfo.TxResult == common.SendSuccess && txInfo.TxHash != "") ||
				(txInfo != nil && txInfo.TxResult == common.SendFailed && txInfo.TxHash != "") {
				boo := mana.VerifyTx(txInfo.TxHash)
				if boo {
					log.Infof("Failed transactions revalidate success, txhash: %s", txInfo.TxHash)
					ti, err := mana.GetTxTime(txInfo.TxHash)
					if err != nil {
						log.Errorf("GetTxTime error: %s", err)
						continue
					}
					err = bonus_db.UpdateTxResult(eventType, param.Address, common.TxSuccess, ti, "")
					if err != nil {
						log.Errorf("UpdateTxResult failed, txhash: %s, error: %s", txInfo.TxHash, err)
					}
					continue
				}
				log.Infof("Failed transactions revalidate failed, txhash: %s", txInfo.TxHash)
			}
			if txInfo != nil && txInfo.TxHex != "" {
				txHex, err = common2.HexToBytes(txInfo.TxHex)
				if err != nil {
					log.Errorf("QueryTxHexByOntid HexToBytes failed, error: %s", err)
					continue
				}
			}
			if txInfo == nil {
				txInfo = &common.TransactionInfo{}
			}
			//build tx
			if txHex == nil {
				var txHash string
				txHash, txHex, err = mana.NewWithdrawTx(param.Address, param.Amount)
				if err != nil || txHash == "" || txHex == nil {
					log.Errorf("Build Transfer Tx failed,address: %s,txHash: %s, err: %s", param.Address, txHash, err)
					if txInfo == nil {
						err := bonus_db.UpdateTxResult(eventType, param.Address, common.BuildTxFailed, 0, err.Error())
						if err != nil {
							log.Errorf("InsertTransactionInfo error: %s, eventType: %s, address: %s", err, eventType, param.Address)
						}
					}
					continue
				}
				//Avoid repetitive insert
				if txInfo == nil {
					err := bonus_db.UpdateTxInfo(txHash, common2.ToHexString(txHex), common.NotSend, eventType, param.Address)
					if err != nil {
						log.Errorf("InsertTransactionInfo error: %s, eventType: %s, address: %s", err, eventType, param.Address)
						continue
					}
				} else {
					err = bonus_db.UpdateTxInfo(txHash, common2.ToHexString(txHex), common.NotSend, eventType, param.Address)
					if err != nil {
						log.Errorf("UpdateTxInfo failed, eventType: %s,txhash: %s, txHex: %s error: %s",
							eventType, txHash, common2.ToHexString(txHex), err)
						continue
					}
				}

				log.Debugf("tx build success, txhash: %s", txHash)
				//update receive txhash
				txInfo.TxHash = txHash
				log.Debugf("InsertTransactionInfo success, eventType: %s, address: %s, txHash: %s",
					eventType, param.Address, txHash)
			}

			log.Debugf("tx build success, txHash: %s, txHex: %s", txInfo.TxHash, common2.ToHexString(txHex))

			//send tx
			retry := 0
			var hash string
			for {
				hash, err = mana.SendTx(txHex)
				if err != nil && retry < config.RetryLimit {
					retry += 1
					time.Sleep(time.Duration(retry*config.SleepTime) * time.Second)
					continue
				} else {
					break
				}
			}
			if err != nil || hash == "" {
				log.Errorf("SendTx error: %s, txhash: %s", err, txInfo.TxHash)
				//save txHex
				err = bonus_db.UpdateTxResult(eventType, param.Address, common.SendFailed, 00, err.Error())
				if err != nil {
					log.Errorf("UpdateTxResult error: %s, eventType: %s, address: %s, txHash: %s, txresult: %d",
						err, eventType, param.Address, txInfo.TxHash, byte(common.SendFailed))
				}
				continue
			} else {
				//save txHex
				err = bonus_db.UpdateTxResult(eventType, param.Address, common.SendSuccess, 0, "")
				if err != nil {
					log.Errorf("UpdateTxResult error: %s, eventType: %s, address: %s, projectId: %d, txHash: %s, txresult: %d",
						err, eventType, param.Address, txInfo.TxHash, byte(common.SendSuccess))
				}
			}
			log.Debugf("tx send success, txhash: %s", txInfo.TxHash)
			self.verifyTxQueue <- hash
		}
	}
}

func (self *TxHandleTask) StartVerifyTxTask(mana interfaces.WithdrawManager) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("recover info: ", r)
		}
	}()
	for {
		select {
		case hash, ok := <-self.verifyTxQueue:
			if !ok || hash == "" {
				close(self.closeChan)
				log.Info("close(self.closeChan), verify over")
				return
			}
			boo := mana.VerifyTx(hash)
			if !boo {
				//save failed tx to bonus_db
				err := bonus_db.UpdateTxResultByTxHash(hash, common.TxFailed, 0)
				if err != nil {
					log.Errorf("UpdateTxResult error: %s, txHash: %s", err, hash)
				}
				log.Errorf("VerifyTx failed, txhash: %s", hash)
				continue
			}
			ti, err := mana.GetTxTime(hash)
			if err != nil {
				log.Errorf("GetTxTime error: %s", err)
				continue
			}
			//update bonus_db
			err = bonus_db.UpdateTxResultByTxHash(hash, common.TxSuccess, ti)
			if err != nil {
				log.Errorf("UpdateTxResult error: %s, txHash: %s", err, hash)
			}
			log.Debugf("verify tx success, txhash: %s", hash)
		}
	}
}
