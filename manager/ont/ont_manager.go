package ont

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ontio/bonus/bonus_db"
	common2 "github.com/ontio/bonus/common"
	"github.com/ontio/bonus/config"
	"github.com/ontio/bonus/manager/transfer"
	"github.com/ontio/bonus/utils"
	sdk "github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology-go-sdk/oep4"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
)

var OntIDVersion = byte(0)

type OntManager struct {
	// persisted
	cfg     *config.Ont
	excel   *common2.ExcelParam
	netType string
	db      *bonus_db.BonusDB

	// loaded
	account         *sdk.Account
	ontSdk          *sdk.OntologySdk
	contractAddress common.Address
	decimals        int
	txHandleTask    *transfer.TxHandleTask
}

func NewOntManager(cfg *config.Ont, eatp *common2.ExcelParam, netType string, db *bonus_db.BonusDB) (*OntManager, error) {
	var rpcAddr string
	if netType == config.MainNet {
		rpcAddr = cfg.OntJsonRpcAddressMainNet
	} else if netType == config.TestNet {
		rpcAddr = cfg.OntJsonRpcAddressTestNet
	} else {
		return nil, fmt.Errorf("NewOntManager not support nettype: %s", netType)
	}
	ontSdk := sdk.NewOntologySdk()
	ontSdk.NewRpcClient().SetAddress(rpcAddr)
	walletPath := config.GetEventDir(eatp.TokenType, eatp.EventType)
	err := common2.CheckPath(walletPath)
	if err != nil {
		return nil, err
	}

	walletName := fmt.Sprintf("%s%s.dat", "ont_", eatp.EventType)
	walletFile := filepath.Join(walletPath, walletName)
	var wallet *sdk.Wallet
	if !common2.PathExists(walletFile) {
		wallet, err = ontSdk.CreateWallet(walletFile)
		if err != nil {
			return nil, err
		}
	} else {
		wallet, err = ontSdk.OpenWallet(walletFile)
		if err != nil {
			return nil, fmt.Errorf("failed to open local wallet %s: %s", walletFile, err)
		}
	}

	log.Infof("ont walletFile: %s", walletFile)
	acct, err := wallet.GetDefaultAccount([]byte(config.PASSWORD))
	if acct == nil {
		acct, err = wallet.NewDefaultSettingAccount([]byte(config.PASSWORD))
		if err != nil {
			return nil, err
		}
		if err := wallet.SetDefaultAccount(acct.Address.ToBase58()); err != nil {
			return nil, err
		}
		err = wallet.Save()
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		log.Errorf("GetDefaultAccount failed, error: %s", err)
		return nil, err
	}
	log.Infof("ont admin address: %s", acct.Address.ToBase58())
	mgr := &OntManager{
		cfg:     cfg,
		excel:   eatp,
		netType: netType,
		db:      db,
		account: acct,
		ontSdk:  ontSdk,
	}

	if mgr.excel.ContractAddress != "" {
		if err := mgr.updateContractInfo(mgr.excel.ContractAddress); err != nil {
			return nil, fmt.Errorf("update contract info: %s", err)
		}
	}

	return mgr, nil
}

type OntPersistHelper struct {
	Config  *config.Ont         `json:"config"`
	Request *common2.ExcelParam `json:"request"`
}

func (this *OntManager) Store() error {
	helper := &OntPersistHelper{
		Config:  this.cfg,
		Request: this.excel,
	}
	data, err := json.Marshal(helper)
	if err != nil {
		return err
	}
	configFilename := filepath.Join(config.GetEventDir(this.excel.TokenType, this.excel.EventType), "config.json")
	if err := ioutil.WriteFile(configFilename, data, 0644); err != nil {
		return err
	}
	return nil
}

func LoadOntManager(tokenType, eventType string) (*config.Ont, *common2.ExcelParam, error) {
	configFilename := filepath.Join(config.GetEventDir(tokenType, eventType), "config.json")
	data, err := ioutil.ReadFile(configFilename)
	if err != nil {
		return nil, nil, err
	}

	helper := &OntPersistHelper{}
	if err := json.Unmarshal(data, helper); err != nil {
		return nil, nil, err
	}
	return helper.Config, helper.Request, nil
}

func (this *OntManager) SetGasPrice(gasPrice uint64) error {
	this.cfg.GasPrice = gasPrice
	// FIXME: update fee-estimate
	return this.Store()
}

func (this *OntManager) GetGasPrice() uint64 {
	return this.cfg.GasPrice
}

func (this *OntManager) QueryTransferProgress() (map[string]int, error) {
	return this.db.QueryTransferProgress(this.excel.EventType, this.excel.NetType)
}

func (this *OntManager) CloseDB() {
	this.db.Close()
}

func (self *OntManager) GetNetType() string {
	return self.netType
}
func (self *OntManager) GetExcelParam() *common2.ExcelParam {
	return self.excel
}

func (self *OntManager) QueryTxInfo(start, end int, txResult common2.TxResult) ([]*common2.TransactionInfo, error) {
	return self.db.QueryTxInfoByEventType(self.excel.EventType, start, end, txResult)
}

func (self *OntManager) VerifyAddress(address string) bool {
	_, err := common.AddressFromBase58(address)
	if err != nil {
		log.Errorf("ont VerifyAddress failed, address: %s, %s", address, err)
		return false
	}
	return true
}

func (self *OntManager) StartTransfer() {
	self.StartHandleTxTask()
	go func() {
		self.txHandleTask.UpdateTxInfoTable(self, self.excel)
		for _, trParam := range self.excel.BillList {
			if trParam.Amount == "0" {
				continue
			}
			select {
			case self.txHandleTask.TransferQueue <- trParam:
			case <-self.txHandleTask.CloseChan:
				break
			}
		}
		// FIXME: remove the close if tx-timeout not handled
		close(self.txHandleTask.TransferQueue)
		self.txHandleTask.WaitClose()
	}()
}

func (self *OntManager) GetStatus() common2.TransferStatus {
	if self.txHandleTask == nil {
		return common2.NotTransfer
	}
	return self.txHandleTask.TransferStatus
}

func (self *OntManager) StartHandleTxTask() {
	txHandleTask := transfer.NewTxHandleTask(self.excel.TokenType, self.db, config.ONT_TRANSFER_QUEUE_SIZE)
	self.txHandleTask = txHandleTask
	log.Infof("init txHandleTask success, transfer status: %d\n", self.txHandleTask.TransferStatus)
	go self.txHandleTask.StartHandleTransferTask(self, self.excel.EventType)
	go self.txHandleTask.StartVerifyTxTask(self)
}

func (self *OntManager) WithdrawToken(address string, tokenType string) error {
	bal, err := self.GetAdminBalance()
	if err != nil {
		return fmt.Errorf("GetAdminBalance faied, error: %s", err)
	}

	// check fee
	fee := new(big.Int).SetUint64(uint64(10000000)) // 0.01
	ongBalance := utils.ToIntByPrecise(bal[config.ONG], config.ONG_DECIMALS)
	if ongBalance.Cmp(fee) < 0 {
		return nil
	}

	var amt string
	if tokenType == config.ONT {
		amt = bal[tokenType]
	} else if tokenType == config.ONG {
		amtBig := new(big.Int).Sub(ongBalance, fee)
		amt = utils.ToStringByPrecise(amtBig, config.ONG_DECIMALS)
	} else if tokenType == config.OEP4 {
		amt = bal[tokenType]
	} else {
		log.Errorf("not support token type: %s", tokenType)
		return fmt.Errorf("not support token type: %s", tokenType)
	}
	return self.withdrawToken(address, tokenType, amt)
}

func (self *OntManager) withdrawToken(address, tokenType, amt string) error {
	log.Infof("address:%s, amt:%s", address, amt)
	_, txHex, err := self.NewWithdrawTx(address, amt, tokenType)
	if err != nil {
		log.Errorf("NewWithdrawTx failed, error: %s", err)
		return fmt.Errorf("NewWithdrawTx failed, error: %s", err)
	}
	hash, err := self.SendTx(txHex)
	if err != nil {
		log.Errorf("SendTx failed,txhash: %s, error: %s", hash, err)
		return fmt.Errorf("SendTx failed,txhash: %s, error: %s", hash, err)
	}
	boo, err := self.VerifyTx(hash, config.RetryLimit)
	if !boo {
		log.Errorf("[withdrawToken] VerifyTx failed,txhash: %s, error: %s", hash, err)
		return fmt.Errorf("VerifyTx failed,txhash: %s, error: %s", hash, err)
	}
	return nil
}

func (self *OntManager) updateContractInfo(address string) error {
	addr, err := common.AddressFromHexString(address)
	if err != nil {
		return err
	}

	//update precision
	preResult, err := self.ontSdk.NeoVM.PreExecInvokeNeoVMContract(addr,
		[]interface{}{"decimals", []interface{}{}})
	if err != nil {
		return err
	}
	res, err := preResult.Result.ToInteger()
	if err != nil {
		return err
	}
	self.decimals = int(res.Int64())
	self.contractAddress = addr
	return nil
}

func (self *OntManager) NewBatchWithdrawTx(addrAndAmts [][]string) (string, []byte, error) {
	var sts []ont.State
	for _, addrAndAmt := range addrAndAmts {
		to, err := common.AddressFromBase58(addrAndAmt[0])
		if err != nil {
			return "", nil, fmt.Errorf("AddressFromBase58 error: %s", err)
		}
		var val *big.Int
		if self.excel.TokenType == config.ONT {
			val = utils.ToIntByPrecise(addrAndAmt[1], config.ONT_DECIMALS)
		} else if self.excel.TokenType == config.ONG {
			val = utils.ToIntByPrecise(addrAndAmt[1], config.ONG_DECIMALS)
		} else {
			log.Errorf("token type not support, tokenType: %s", self.excel.TokenType)
			return "", nil, fmt.Errorf("not supprt token type: %s", self.excel.TokenType)
		}
		st := ont.State{
			From:  self.account.Address,
			To:    to,
			Value: val.Uint64(),
		}
		sts = append(sts, st)
	}
	params := ont.Transfers{
		States: sts,
	}
	tx, err := self.ontSdk.Native.NewNativeInvokeTransaction(self.cfg.GasPrice, self.cfg.GasLimit,
		OntIDVersion, sdk.ONG_CONTRACT_ADDRESS, "transfer", []interface{}{params})
	if err != nil {
		return "", nil, fmt.Errorf("transfer ong, this.ontologySdk.Native.NewNativeInvokeTransaction error: %s", err)
	}
	err = self.ontSdk.SignToTransaction(tx, self.account)
	if err != nil {
		return "", nil, fmt.Errorf("transfer ong, this.ontologySdk.SignToTransaction err: %s", err)
	}
	t, err := tx.IntoImmutable()
	if err != nil {
		return "", nil, fmt.Errorf("IntoImmutable error: %s", err)
	}
	h := tx.Hash()
	return h.ToHexString(), t.ToArray(), nil
}

func (self *OntManager) NewWithdrawTx(destAddr, amount, tokenType string) (string, []byte, error) {
	address, err := common.AddressFromBase58(destAddr)
	if err != nil {
		return "", nil, fmt.Errorf("common.AddressFromBase58 error: %s", err)
	}
	var tx *types.MutableTransaction
	if (self.excel.TokenType == config.ONT && tokenType == "") || tokenType == config.ONT {
		value := utils.ParseAssetAmount(amount, config.ONT_DECIMALS)
		var sts []ont.State
		sts = append(sts, ont.State{
			From:  self.account.Address,
			To:    address,
			Value: value,
		})
		params := ont.Transfers{
			States: sts,
		}
		tx, err = self.ontSdk.Native.NewNativeInvokeTransaction(self.cfg.GasPrice, self.cfg.GasLimit,
			OntIDVersion, sdk.ONT_CONTRACT_ADDRESS, "transfer", []interface{}{params})
		if err != nil {
			return "", nil, fmt.Errorf("transfer ont, this.ontologySdk.Native.NewNativeInvokeTransaction error: %s", err)
		}
		err = self.ontSdk.SignToTransaction(tx, self.account)
		if err != nil {
			return "", nil, fmt.Errorf("transfer ont: this.ontologySdk.SignToTransaction err: %s", err)
		}
	} else if (self.excel.TokenType == config.ONG && tokenType == "") || tokenType == config.ONG {
		value := utils.ParseAssetAmount(amount, config.ONG_DECIMALS)
		var sts []ont.State
		sts = append(sts, ont.State{
			From:  self.account.Address,
			To:    address,
			Value: value,
		})
		params := ont.Transfers{
			States: sts,
		}
		tx, err = self.ontSdk.Native.NewNativeInvokeTransaction(self.cfg.GasPrice, self.cfg.GasLimit,
			OntIDVersion, sdk.ONG_CONTRACT_ADDRESS, "transfer", []interface{}{params})
		if err != nil {
			return "", nil, fmt.Errorf("transfer ong, this.ontologySdk.Native.NewNativeInvokeTransaction error: %s", err)
		}
		err = self.ontSdk.SignToTransaction(tx, self.account)
		if err != nil {
			return "", nil, fmt.Errorf("transfer ong, this.ontologySdk.SignToTransaction err: %s", err)
		}
	} else if (self.excel.TokenType == config.OEP4 && tokenType == "") || tokenType == config.OEP4 {
		if self.contractAddress == common.ADDRESS_EMPTY {
			return "", nil, fmt.Errorf("contractAddress is nil")
		}
		val := utils.ParseAssetAmount(amount, self.decimals)
		value := new(big.Int).SetUint64(val)
		tx, err = self.ontSdk.NeoVM.NewNeoVMInvokeTransaction(self.cfg.GasPrice, self.cfg.GasLimit, self.contractAddress, []interface{}{"transfer", []interface{}{self.account.Address, address, value}})
		if err != nil {
			return "", nil, fmt.Errorf("NewNeoVMInvokeTransaction error: %s", err)
		}
		err = self.ontSdk.SignToTransaction(tx, self.account)
		if err != nil {
			return "", nil, fmt.Errorf("OEP4 SignToTransaction error: %s", err)
		}
	} else {
		return "", nil, fmt.Errorf("[NewWithdrawTx] not supprt self.eatp.TokenType: %s,token Type: %s", self.excel.TokenType, tokenType)
	}
	t, err := tx.IntoImmutable()
	if err != nil {
		return "", nil, fmt.Errorf("IntoImmutable error: %s", err)
	}
	h := tx.Hash()
	return h.ToHexString(), t.ToArray(), nil
}

func (self *OntManager) GetAdminAddress() string {
	return self.account.Address.ToBase58()
}

func (self *OntManager) GetAdminBalance() (map[string]string, error) {
	res := make(map[string]string)

	// get ONG balance
	val, err := self.ontSdk.Native.Ong.BalanceOf(self.account.Address)
	if err != nil {
		return nil, err
	}
	r := new(big.Int)
	r.SetUint64(val)
	ongBa := utils.ToStringByPrecise(r, 9)
	res[config.ONG] = ongBa

	if self.excel.TokenType == config.ONT {
		val, err := self.ontSdk.Native.Ont.BalanceOf(self.account.Address)
		if err != nil {
			return nil, err
		}
		ba := strconv.FormatUint(val, 10)
		res[config.ONT] = ba
	} else if self.excel.TokenType == config.OEP4 {
		oep4 := oep4.NewOep4(self.contractAddress, self.ontSdk)
		val, err := oep4.BalanceOf(self.account.Address)
		if err != nil {
			return nil, err
		}
		ba := utils.ToStringByPrecise(val, uint64(self.decimals))
		res[self.excel.TokenType] = ba
	}
	return res, nil
}

func (self *OntManager) EstimateFee(tokenType string, total int) (string, error) {
	fee := float64(total) * 0.01
	return strconv.FormatFloat(fee, 'f', -1, 64), nil
}

func (self *OntManager) GetTotal() int {
	return len(self.excel.BillList)
}

func (self *OntManager) ComputeSum() (string, error) {
	sum := uint64(0)
	if self.excel.TokenType == config.ONT {
		for _, item := range self.excel.BillList {
			val, err := strconv.ParseUint(item.Amount, 10, 64)
			if err != nil {
				return "", err
			}
			sum += val
		}
		return strconv.FormatUint(sum, 10), nil
	} else if self.excel.TokenType == config.ONG {
		for _, item := range self.excel.BillList {
			val := utils.ParseAssetAmount(item.Amount, 9)
			sum += val
		}
		temp := new(big.Int)
		temp.SetUint64(sum)
		return utils.ToStringByPrecise(temp, uint64(9)), nil
	} else if self.excel.TokenType == config.OEP4 {
		for _, item := range self.excel.BillList {
			val := utils.ParseAssetAmount(item.Amount, self.decimals)
			sum += val
		}
		temp := new(big.Int)
		temp.SetUint64(sum)
		return utils.ToStringByPrecise(temp, uint64(self.decimals)), nil
	} else {
		return "", fmt.Errorf("not support token type: %s", self.excel.TokenType)
	}
}

func (self *OntManager) GetTxTime(txHash string) (uint32, error) {
	height, err := self.ontSdk.GetBlockHeightByTxHash(txHash)
	if err != nil {
		return 0, err
	}
	block, err := self.ontSdk.GetBlockByHeight(height)
	if err != nil {
		return 0, err
	}
	return block.Header.Timestamp, nil
}
func (self *OntManager) SendTx(txHex []byte) (string, error) {
	tx, err := types.TransactionFromRawBytes(txHex)
	if err != nil {
		return "", fmt.Errorf("TransactionFromRawBytes error: %s", err)
	}
	txMu, err := tx.IntoMutable()
	if err != nil {
		return "", fmt.Errorf("IntoMutable error: %s", err)
	}
	txHash, err := self.ontSdk.SendTransaction(txMu)
	if err != nil {
		return "", fmt.Errorf("SendTransaction error: %s", err)
	}
	return txHash.ToHexString(), nil
}

func (self *OntManager) VerifyTx(txHash string, retryLimit int) (bool, error) {
	retry := 0
	for {
		event, err := self.ontSdk.GetSmartContractEvent(txHash)
		if event != nil && event.State == 0 {
			return false, fmt.Errorf("tx failed")
		}
		if err != nil && retry < retryLimit {
			if err != nil {
				log.Errorf("GetSmartContractEvent error: %s, retry: %d, txHash: %s", err, retry, txHash)
			}
			retry += 1
			time.Sleep(time.Duration(retry*config.SleepTime) * time.Second)
			continue
		}
		if err != nil && retry >= retryLimit {
			log.Errorf("GetSmartContractEvent fail, txhash: %s, err: %s", txHash, err)
			return false, err
		}
		if event == nil && retry < retryLimit {
			retry += 1
			time.Sleep(time.Duration(retry*config.SleepTime) * time.Second)
			continue
		}
		if event == nil {
			return false, fmt.Errorf("no the transaction")
		}
		if event.State == 1 {
			if self.contractAddress != common.ADDRESS_EMPTY {
				if len(event.Notify) == 2 {
					return true, nil
				} else {
					return false, fmt.Errorf("uncertain tx")
				}
			}
			return true, nil
		}
		return false, fmt.Errorf("tx failed")
	}
}
