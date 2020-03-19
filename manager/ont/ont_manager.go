package ont

import (
	"fmt"
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
	"math/big"
	"os"
	"strconv"
	"time"
)

var OntIDVersion = byte(0)

type OntManager struct {
	account         *sdk.Account
	ontSdk          *sdk.OntologySdk
	contractAddress common.Address
	cfg             *config.Ont
	ipIndex         int
	precision       int
	txHandleTask    *transfer.TxHandleTask
	eatp            *common2.ExcelParam
	netType         string
}

func NewOntManager(cfg *config.Ont, eatp *common2.ExcelParam, netType string) (*OntManager, error) {
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
	if cfg.WalletFile == "" {
		cfg.WalletFile = fmt.Sprintf("%s%s%s", config.DefaultWalletPath, string(os.PathSeparator), "ont")
	}
	err := common2.CheckPath(cfg.WalletFile)
	if err != nil {
		return nil, err
	}

	walletName := fmt.Sprintf("%s%s.dat", "ont_", eatp.EventType)
	walletFile := fmt.Sprintf("%s%s%s", cfg.WalletFile, string(os.PathSeparator), walletName)
	var wallet *sdk.Wallet
	if !common2.PathExists(walletFile) {
		wallet, err = ontSdk.CreateWallet(walletFile)
		if err != nil {
			return nil, err
		}
	} else {
		wallet, err = ontSdk.OpenWallet(walletFile)
		if err != nil {
			log.Fatalf("Can't open local wallet: %s", err)
			return nil, fmt.Errorf("password is wrong")
		}
	}

	log.Infof("ont walletFile: %s", walletFile)
	acct, err := wallet.GetDefaultAccount([]byte(config.PASSWORD))
	if (err != nil && err.Error() == "does not set default account") || acct == nil {
		acct, err = wallet.NewDefaultSettingAccount([]byte(config.PASSWORD))
		if err != nil {
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

	ontManager := &OntManager{
		account: acct,
		ontSdk:  ontSdk,
		cfg:     cfg,
		eatp:    eatp,
		netType: netType,
	}

	return ontManager, nil
}

func (self *OntManager) GetNetType() string {
	return self.netType
}
func (self *OntManager) GetExcelParam() *common2.ExcelParam {
	return self.eatp
}

func (self *OntManager) VerifyAddress(address string) bool {
	_, err := common.AddressFromBase58(address)
	if err != nil {
		log.Errorf("ont VerifyAddress failed, address: %s", address)
		return false
	}
	return true
}

func (self *OntManager) StartTransfer() {
	self.StartHandleTxTask()
	go func() {
		self.txHandleTask.UpdateTxInfoTable(self, self.eatp)
		for _, trParam := range self.eatp.BillList {
			if trParam.Amount == "0" {
				continue
			}
			self.txHandleTask.TransferQueue <- trParam
		}
		close(self.txHandleTask.TransferQueue)
		self.txHandleTask.WaitClose()
	}()
}

func (self *OntManager) GetStatus() common2.TransferStatus {
	if self.txHandleTask == nil {
		log.Info("self.txHandleTask is nil")
		return common2.NotTransfer
	}
	return self.txHandleTask.TransferStatus
}

func (self *OntManager) StartHandleTxTask() {
	txHandleTask := transfer.NewTxHandleTask(self.eatp.TokenType)
	self.txHandleTask = txHandleTask
	log.Infof("init txHandleTask success, transfer status: %d\n", self.txHandleTask.TransferStatus)
	go self.txHandleTask.StartHandleTransferTask(self, self.eatp.EventType)
	go self.txHandleTask.StartVerifyTxTask(self)
}

func (self *OntManager) WithdrawToken(address string, tokenType string) error {
	bal, err := self.GetAdminBalance()
	if err != nil {
		return fmt.Errorf("GetAdminBalance faied, error: %s", err)
	}
	var amt string
	b := bal[tokenType]
	if tokenType == config.ONT {
		amt = b
	} else if tokenType == config.ONG {
		bigInt := utils.ToIntByPrecise(b, config.ONG_DECIMALS)
		fee := new(big.Int).SetUint64(uint64(10000000))
		amtBig := new(big.Int).Sub(bigInt, fee)
		amt = utils.ToStringByPrecise(amtBig, config.ONG_DECIMALS)
	} else if tokenType == config.OEP4 {
		amt = b
	} else {
		log.Errorf("not support token type: %s", tokenType)
		return fmt.Errorf("not support token type: %s", tokenType)
	}
	self.withdrawToken(address, tokenType, amt)
	return nil
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

func (self *OntManager) SetContractAddress(address string) error {
	addr, err := common.AddressFromHexString(address)
	if err != nil {
		return err
	}
	self.contractAddress = addr
	if self.eatp.TokenType == config.OEP5 {
		return nil
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
	self.precision = int(res.Int64())
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
		if self.eatp.TokenType == config.ONT {
			val = utils.ToIntByPrecise(addrAndAmt[1], config.ONT_DECIMALS)
		} else if self.eatp.TokenType == config.ONG {
			val = utils.ToIntByPrecise(addrAndAmt[1], config.ONG_DECIMALS)
		} else {
			log.Errorf("token type not support, tokenType: %s", self.eatp.TokenType)
			return "", nil, fmt.Errorf("not supprt token type: %s", self.eatp.TokenType)
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
	if (self.eatp.TokenType == config.ONT && tokenType == "") || tokenType == config.ONT {
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
	} else if (self.eatp.TokenType == config.ONG && tokenType == "") || tokenType == config.ONG {
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
	} else if (self.eatp.TokenType == config.OEP4 && tokenType == "") || tokenType == config.OEP4 {
		if self.contractAddress == common.ADDRESS_EMPTY {
			return "", nil, fmt.Errorf("contractAddress is nil")
		}
		val := utils.ParseAssetAmount(amount, self.precision)
		value := new(big.Int).SetUint64(val)
		tx, err = self.ontSdk.NeoVM.NewNeoVMInvokeTransaction(self.cfg.GasPrice, self.cfg.GasLimit, self.contractAddress, []interface{}{"transfer", []interface{}{self.account.Address, address, value}})
		if err != nil {
			return "", nil, fmt.Errorf("NewNeoVMInvokeTransaction error: %s", err)
		}
		err = self.ontSdk.SignToTransaction(tx, self.account)
		if err != nil {
			return "", nil, fmt.Errorf("OEP4 SignToTransaction error: %s", err)
		}
	} else if (self.eatp.TokenType == config.OEP5 && tokenType == "") || tokenType == config.OEP5 {
		if self.contractAddress == common.ADDRESS_EMPTY {
			return "", nil, fmt.Errorf("contractAddress is nil")
		}
		tokenId := ""
		tx, err = self.ontSdk.NeoVM.NewNeoVMInvokeTransaction(self.cfg.GasPrice, self.cfg.GasLimit, self.contractAddress, []interface{}{"transfer", []interface{}{address, tokenId}})
		if err != nil {
			return "", nil, fmt.Errorf("NewNeoVMInvokeTransaction error: %s", err)
		}
		err = self.ontSdk.SignToTransaction(tx, self.account)
		if err != nil {
			return "", nil, fmt.Errorf("OEP4 SignToTransaction error: %s", err)
		}
	} else {
		return "", nil, fmt.Errorf("[NewWithdrawTx] not supprt self.eatp.TokenType: %s,token Type: %s", self.eatp.TokenType, tokenType)
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
	val, err := self.ontSdk.Native.Ont.BalanceOf(self.account.Address)
	if err != nil {
		return nil, err
	}
	ontBa := strconv.FormatUint(val, 10)
	val, err = self.ontSdk.Native.Ong.BalanceOf(self.account.Address)
	if err != nil {
		return nil, err
	}
	r := new(big.Int)
	r.SetUint64(val)
	ongBa := utils.ToStringByPrecise(r, 9)
	res := make(map[string]string)
	res[config.ONT] = ontBa
	res[config.ONG] = ongBa
	if self.eatp.TokenType == config.OEP4 {
		oep4 := oep4.NewOep4(self.contractAddress, self.ontSdk)
		val, err := oep4.BalanceOf(self.account.Address)
		if err != nil {
			return nil, err
		}
		ba := utils.ToStringByPrecise(val, uint64(self.precision))
		res[self.eatp.TokenType] = ba
	}
	return res, nil
}

func (self *OntManager) EstimateFee(tokenType string, total int) (string, error) {
	fee := float64(total) * 0.01
	return strconv.FormatFloat(fee, 'f', -1, 64), nil
}

func (self *OntManager) GetTotal() int {
	return len(self.eatp.BillList)
}

func (self *OntManager) ComputeSum() (string, error) {
	sum := uint64(0)
	if self.eatp.TokenType == config.ONT {
		for _, item := range self.eatp.BillList {
			val, err := strconv.ParseUint(item.Amount, 10, 64)
			if err != nil {
				return "", err
			}
			sum += val
		}
		return strconv.FormatUint(sum, 10), nil
	} else if self.eatp.TokenType == config.ONG {
		for _, item := range self.eatp.BillList {
			val := utils.ParseAssetAmount(item.Amount, 9)
			sum += val
		}
		temp := new(big.Int)
		temp.SetUint64(sum)
		return utils.ToStringByPrecise(temp, uint64(9)), nil
	} else if self.eatp.TokenType == config.OEP4 {
		for _, item := range self.eatp.BillList {
			val := utils.ParseAssetAmount(item.Amount, self.precision)
			sum += val
		}
		temp := new(big.Int)
		temp.SetUint64(sum)
		return utils.ToStringByPrecise(temp, uint64(self.precision)), nil
	} else {
		return "", fmt.Errorf("not support token type: %s", self.eatp.TokenType)
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
