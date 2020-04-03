package eth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	ethComm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ontio/bonus/bonus_db"
	common2 "github.com/ontio/bonus/common"
	"github.com/ontio/bonus/config"
	"github.com/ontio/bonus/manager/transfer"
	"github.com/ontio/bonus/utils"
	common3 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
)

var (
	OneGwei           = new(big.Int).SetUint64(uint64(1000000000))
	DEFAULT_GAS_PRICE = utils.ToIntByPrecise("0.00000004", config.ETH_DECIMALS) // 40 Gwei
	MIN_ETH_BANALNCE  = utils.ToIntByPrecise("0.00001", config.ETH_DECIMALS)
)

type Token struct {
	ContractAddr ethComm.Address
	Contract     *Erc20
	Decimals     uint64
}

type EthManager struct {
	// persisted
	cfg     *config.Eth
	excel   *common2.ExcelParam
	netType string
	db      *bonus_db.BonusDB

	keyStore     *keystore.KeyStore
	account      accounts.Account
	ethClient    *ethclient.Client
	txTimeClient *rpc.Client
	tokens       map[string]*Token
	Erc20Abi     abi.ABI
	nonce        uint64
	txHandleTask *transfer.TxHandleTask
	stopChan     chan bool
}

func NewEthManager(cfg *config.Eth, eatp *common2.ExcelParam, netType string) (*EthManager, error) {
	var rpcAddr string
	if netType == config.MainNet {
		rpcAddr = cfg.RpcAddrMainNet
	} else if netType == config.TestNet {
		rpcAddr = cfg.RpcAddrTestNet
	} else {
		return nil, fmt.Errorf("[NewEthManager] not support net type: %s", netType)
	}
	if rpcAddr == "" {
		return nil, fmt.Errorf("RpcAddr config error")
	}

	ethClient, err := ethclient.Dial(rpcAddr)
	if err != nil {
		return nil, fmt.Errorf("NewEthManager: connect to node failed, %s", err)
	}
	c, err := rpc.DialContext(context.Background(), rpcAddr)
	if err != nil {
		return nil, err
	}
	walletDir := common2.GetEventDir(eatp.TokenType, eatp.EventType)
	err = common2.CheckDir(walletDir)
	if err != nil {
		return nil, err
	}

	walletPath := filepath.Join(walletDir, "eth_"+eatp.EventType)
	log.Infof("eth wallet path: %s", walletPath)

	keyStore := keystore.NewKeyStore(walletPath, keystore.StandardScryptN, keystore.StandardScryptP)
	accs := keyStore.Accounts()
	var account accounts.Account
	if len(accs) == 0 {
		account, err = keyStore.NewAccount(config.PASSWORD)
		if err != nil {
			return nil, err
		}
	} else {
		account = accs[0]
	}
	log.Infof("eth admin address: %s", account.Address.Hex())

	mgr := &EthManager{
		tokens:       make(map[string]*Token),
		account:      account,
		keyStore:     keyStore,
		cfg:          cfg,
		txTimeClient: c,
		excel:        eatp,
		netType:      netType,
		stopChan:     make(chan bool),
	}

	nonce, err := ethClient.PendingNonceAt(context.Background(), account.Address)
	if err != nil {
		return nil, fmt.Errorf("NewEthManager: fetch nonce failed, %s", err)
	}
	mgr.nonce = nonce
	mgr.ethClient = ethClient

	mgr.Erc20Abi, err = abi.JSON(strings.NewReader(Erc20ABI))
	if err != nil {
		return nil, fmt.Errorf("NewEthManager: parse erc20 abi failed, %s", err)
	}

	if mgr.excel.ContractAddress != "" {
		if err := mgr.updateContractInfo(mgr.excel.ContractAddress); err != nil {
			return nil, fmt.Errorf("update contract info: %s", err)
		}
	}
	return mgr, nil
}
func (this *EthManager) SetDB(db *bonus_db.BonusDB) {
	this.db = db
}

func parseGasPriceToGwei(gasPrice *big.Int) int {
	b := new(big.Int).Div(gasPrice, OneGwei)
	return int(b.Int64())
}
func parseGweiToGasPrice(gasPrice int) *big.Int {
	temp := new(big.Int).SetUint64(uint64(gasPrice))
	return new(big.Int).Mul(temp, OneGwei)
}

type EthPersistHelper struct {
	Config  *config.Eth         `json:"config"`
	Request *common2.ExcelParam `json:"request"`
}

func (this *EthManager) Store() error {
	for _, item := range this.excel.BillList {
		if !this.VerifyAddress(item.Address) {
			return fmt.Errorf("invalid address: %s", item.Address)
		}
	}
	helper := &EthPersistHelper{
		Config:  this.cfg,
		Request: this.excel,
	}
	data, err := json.Marshal(helper)
	if err != nil {
		return err
	}
	configFilename := filepath.Join(common2.GetEventDir(this.excel.TokenType, this.excel.EventType), "config.json")
	if err := ioutil.WriteFile(configFilename, data, 0644); err != nil {
		return err
	}
	return nil
}

func LoadEthManager(tokenType, eventType string) (*config.Eth, *common2.ExcelParam, error) {
	configFilename := filepath.Join(common2.GetEventDir(tokenType, eventType), "config.json")
	data, err := ioutil.ReadFile(configFilename)
	if err != nil {
		return nil, nil, err
	}

	helper := &EthPersistHelper{}
	if err := json.Unmarshal(data, helper); err != nil {
		return nil, nil, err
	}
	helper.Request.TrParamSort()
	return helper.Config, helper.Request, nil
}

func (this *EthManager) QueryTransferProgress() (map[string]int, error) {
	return this.db.QueryTransferProgress(this.excel.EventType, this.excel.NetType)
}

func (this *EthManager) CloseDB() {
	this.db.Close()
}
func (this *EthManager) SetGasPrice(gasPrice uint64) error {
	this.cfg.GasPrice = gasPrice
	// FIXME: update fee-estimate
	return this.Store()
}

func (this *EthManager) GetGasPrice() uint64 {
	return this.cfg.GasPrice
}

func (self *EthManager) GetNetType() string {
	return self.netType
}

func (this *EthManager) GetExcelParam() *common2.ExcelParam {
	return this.excel
}

func (self *EthManager) QueryTxInfo(start, end int, txResult common2.TxResult) ([]*common2.TransactionInfo, int, error) {
	return self.db.QueryTxInfoByEventType(self.excel.EventType, start, end, txResult)
}

func (self *EthManager) VerifyAddress(address string) bool {
	boo := ethComm.IsHexAddress(address)
	if !boo {
		log.Errorf("eth VerifyAddress failed, address: %s", address)
		return false
	}
	return true
}

func (self *EthManager) GetAdminAddress() string {
	return self.account.Address.Hex()
}

func (self *EthManager) GetAdminBalance() (map[string]string, error) {
	res := make(map[string]string)
	if self.excel.TokenType == config.ERC20 {
		erc20, ok := self.tokens[self.excel.ContractAddress]
		if !ok || erc20 == nil {
			return nil, fmt.Errorf("Withdraw: token %s not exist", self.excel.ContractAddress)
		}
		balance, err := erc20.Contract.BalanceOf(&bind.CallOpts{Pending: false}, self.account.Address)
		if err != nil {
			return nil, fmt.Errorf("Withdraw: cannot get self balance, token %s, err: %s", self.excel.ContractAddress, err)
		}
		res[self.excel.TokenType] = utils.ToStringByPrecise(balance, erc20.Decimals)
	}
	ethBalance, err := self.ethClient.BalanceAt(context.Background(), self.account.Address, nil)
	if err != nil {
		return nil, fmt.Errorf("[GetAdminBalance] Withdraw: cannot get eth pending balance, err: %s", err)
	}
	res[config.ETH] = utils.ToStringByPrecise(ethBalance, config.ETH_DECIMALS)
	return res, nil
}

func (self *EthManager) EstimateFee(tokenType string, total int) (string, error) {
	contractAddr := ethComm.HexToAddress(self.excel.ContractAddress)
	adminAddr := self.GetAdminAddress()
	adminAddress := ethComm.HexToAddress(adminAddr)
	// TODO: FIX HERE
	amount := utils.ToIntByPrecise("2000", config.ETH_DECIMALS)
	gaslimit, err := self.estimateGasLimit(tokenType, contractAddr, adminAddress, amount, DEFAULT_GAS_PRICE)
	if err != nil {
		return "", err
	}
	if total > 1 {
		gaslimit = gaslimit * 2
	}
	gasLimi := new(big.Int).SetUint64(gaslimit)
	gas := new(big.Int).Mul(gasLimi, DEFAULT_GAS_PRICE)
	gasTotal := new(big.Int).Mul(gas, new(big.Int).SetUint64(uint64(total)))
	return utils.ToStringByPrecise(gasTotal, config.ETH_DECIMALS), nil
}

func (this *EthManager) GetTotal() int {
	return len(this.excel.BillList)
}

func (self *EthManager) ComputeSum() (string, error) {
	// TODO: SUPPORT ETH
	if self.excel.TokenType == config.ERC20 {
		sum := new(big.Int)
		for _, item := range self.excel.BillList {
			val := utils.ToIntByPrecise(item.Amount, self.tokens[self.excel.ContractAddress].Decimals)
			sum = sum.Add(sum, val)
		}
		return utils.ToStringByPrecise(sum, self.tokens[self.excel.ContractAddress].Decimals), nil
	}
	return "", fmt.Errorf("[ComputeSum]not supported token type: %s", self.excel.TokenType)
}

func (self *EthManager) WithdrawToken(address, tokenType string) error {
	allBalances, err := self.GetAdminBalance()
	if err != nil {
		return err
	}
	var amt string
	if tokenType == config.ERC20 {
		amt = allBalances[config.ERC20]
		if amt == "" {
			return fmt.Errorf("erc20 balance is 0")
		}
		feeStr, err := self.EstimateFee(config.ERC20, 1)
		if err != nil {
			return err
		}
		fee := utils.ToIntByPrecise(feeStr, config.ETH_DECIMALS)

		ethBalance := allBalances[config.ETH]
		ethBa := utils.ToIntByPrecise(ethBalance, config.ETH_DECIMALS)
		if ethBa.Cmp(fee) <= 0 {
			return fmt.Errorf("not enough eth")
		}
	} else if tokenType == config.ETH {
		ba := allBalances[config.ETH]
		if ba == "" {
			return fmt.Errorf("eth balance is 0")
		}
		baBig := utils.ToIntByPrecise(ba, config.ETH_DECIMALS)
		feeStr, err := self.EstimateFee(config.ETH, 1)
		if err != nil {
			return err
		}
		fee := utils.ToIntByPrecise(feeStr, config.ETH_DECIMALS)
		if baBig.Cmp(fee) <= 0 {
			return fmt.Errorf("not enough eth")
		}
		amtBig := new(big.Int).Sub(baBig, fee)
		amt = utils.ToStringByPrecise(amtBig, config.ETH_DECIMALS)
	}
	log.Infof("[WithdrawToken] amt: %s, tokenType:%s", amt, tokenType)
	nonce, err := self.ethClient.PendingNonceAt(context.Background(), self.account.Address)
	if err != nil {
		return fmt.Errorf("[WithdrawToken] fetch nonce failed, %s", err)
	}
	self.nonce = nonce
	hash, txHex, err := self.NewWithdrawTx(address, amt, tokenType)
	if hash == "" || txHex == nil || err != nil {
		return fmt.Errorf("[WithdrawToken] NewWithdrawTx failed, error: %s", err)
	}
	hash, err = self.SendTx(txHex)
	if err != nil {
		return fmt.Errorf("send tx failed, error:%s", err)
	}
	boo, err := self.VerifyTx(hash, config.RetryLimit)
	if !boo {
		return fmt.Errorf("verify tx failed, error:%s, hash:%s", err, hash)
	}
	return nil
}

func (this *EthManager) updateContractInfo(address string) error {
	contractAddr := ethComm.HexToAddress(address)
	erc20, err := NewErc20(contractAddr, this.ethClient)
	if err != nil {
		return fmt.Errorf("NewEthManager: new erc20 contract failed, token %s, err: %s", this.excel.ContractAddress, err)
	}
	decimals, err := erc20.Decimals(&bind.CallOpts{})
	if err != nil {
		return fmt.Errorf("NewEthManager: cannot get %s decimals, err: %s", this.excel.ContractAddress, err)
	}
	this.tokens[this.excel.ContractAddress] = &Token{ContractAddr: contractAddr, Decimals: decimals.Uint64(), Contract: erc20}
	return nil
}

func (self *EthManager) hasEnoughBalance(amount string) error {
	if self.excel.TokenType == config.ETH {
		amt := utils.ToIntByPrecise(amount, config.ETH_DECIMALS)
		balance, err := self.ethClient.PendingBalanceAt(context.Background(), self.account.Address)
		if err != nil {
			log.Errorf("[StartTransfer] PendingBalanceAt failed: %s", err)
			return fmt.Errorf("[StartTransfer] PendingBalanceAt failed: %s", err)
		}
		baStr := utils.ToStringByPrecise(balance, config.ETH_DECIMALS)
		if balance.Cmp(amt) < 0 {
			return fmt.Errorf("[StartTransfer] not enough balance, balance: %s, amt: %s", baStr, amount)
		}
	} else if self.excel.TokenType == config.ERC20 {
		erc20, ok := self.tokens[self.excel.ContractAddress]
		if !ok || erc20 == nil {
			return fmt.Errorf("[StartTransfer]Withdraw: token %s not exist", self.excel.ContractAddress)
		}
		amt := utils.ToIntByPrecise(amount, erc20.Decimals)
		balance, err := erc20.Contract.BalanceOf(&bind.CallOpts{Pending: false}, self.account.Address)
		if err != nil {
			return fmt.Errorf("[StartTransfer] Withdraw: cannot get self balance, token %s, err: %s", self.excel.ContractAddress, err)
		}
		if balance.Cmp(amt) < 0 {
			baStr := utils.ToStringByPrecise(balance, erc20.Decimals)
			return fmt.Errorf("[StartTransfer] not enough balance, balance: %s, amt: %s", baStr, amount)
		}
	}
	return nil
}

func (self *EthManager) Stop() {
	if self.txHandleTask.TransferStatus != common2.Transfering {
		return
	}
	if self.txHandleTask == nil {
		return
	}
	self.stopChan <- true
	self.txHandleTask.TransferStatus = common2.Stop
}

func (self *EthManager) StartTransfer() {
	self.StartHandleTxTask()
	go func() {
		hasBuildTx, err := self.txHandleTask.UpdateTxInfoTable(self, self.excel)
		if err != nil {
			log.Errorf("[StartTransfer] UpdateTxInfoTable error: %s", err)
			close(self.txHandleTask.TransferQueue)
			self.txHandleTask.WaitClose()
			return
		}
	loop:
		for _, trParam := range self.excel.BillList {
			if trParam.Amount == "0" || trParam.Amount == "" {
				continue
			}
			if !hasBuildTx[trParam.Id] {
				if err = self.hasEnoughBalance(trParam.Amount); err != nil {
					log.Errorf("[StartTransfer]hasEnoughBalance error: %s, id: %d", err, trParam.Id)
					break loop
				}
			}
			select {
			case self.txHandleTask.TransferQueue <- trParam:
				log.Infof("[StartTransfer]TransferQueue id: %d", trParam.Id)
			case <-self.txHandleTask.CloseChan:
				log.Infof("[StartTransfer] CloseChan, id: %d", trParam.Id)
				break loop
			case <-self.stopChan:
				log.Infof("[StartTransfer] stop, id: %d", trParam.Id)
				self.txHandleTask.StopTransferChan <- true
				return
			}
		}
		close(self.txHandleTask.TransferQueue)
		self.txHandleTask.WaitClose()
	}()
}

func (self *EthManager) GetStatus() common2.TransferStatus {
	if self.txHandleTask == nil {
		return common2.NotTransfer
	}
	return self.txHandleTask.TransferStatus
}

func (self *EthManager) StartHandleTxTask() {
	//start transfer task and verify task
	self.txHandleTask = transfer.NewTxHandleTask(self.excel.TokenType, self.db, config.ETH_TRANSFER_QUEUE_SIZE)
	go self.txHandleTask.StartHandleTransferTask(self, self.excel.EventType)
	go self.txHandleTask.StartVerifyTxTask(self)
}

func (this *EthManager) NewWithdrawTx(destAddr, amount, tokenType string) (string, []byte, error) {
	to := ethComm.Address{}
	if ethComm.IsHexAddress(destAddr) {
		to = ethComm.HexToAddress(destAddr)
	} else {
		return "", nil, fmt.Errorf("Withdraw: dest addr is invaild")
	}
	ethBalance, err := this.ethClient.PendingBalanceAt(context.Background(), this.account.Address)
	if err != nil {
		return "", nil, fmt.Errorf("Withdraw: cannot get eth pending balance, err: %s", err)
	}
	if ethBalance.Cmp(MIN_ETH_BANALNCE) < 0 {
		return "", nil, fmt.Errorf("Withdraw: self eth pending balance %s not enough",
			utils.ToStringByPrecise(ethBalance, config.ETH_DECIMALS))
	}
	if (this.excel.TokenType == config.ETH && tokenType == "") || tokenType == config.ETH {
		withdrawAmount := utils.ToIntByPrecise(amount, config.ETH_DECIMALS)
		if ethBalance.Cmp(withdrawAmount) < 0 {
			return "", nil, fmt.Errorf("%s", config.InSufficientBalance)
		}
		log.Debugf("Withdraw: %s, pending balance is %s", this.excel.TokenType,
			utils.ToStringByPrecise(ethBalance, config.ETH_DECIMALS))
		return this.newWithdrawEthTx(to, withdrawAmount, DEFAULT_GAS_PRICE)
	} else {
		erc20, ok := this.tokens[this.excel.ContractAddress]
		if !ok {
			return "", nil, fmt.Errorf("Withdraw: token %s not exist", this.excel.ContractAddress)
		}
		withdrawAmount := utils.ToIntByPrecise(amount, erc20.Decimals)
		balance, err := erc20.Contract.BalanceOf(&bind.CallOpts{Pending: false}, this.account.Address)
		if err != nil {
			return "", nil, fmt.Errorf("Withdraw: cannot get self balance, token %s, err: %s", this.excel.ContractAddress, err)
		}
		if balance.Cmp(withdrawAmount) < 0 {
			return "", nil, fmt.Errorf("%s", config.InSufficientBalance)
		}
		log.Debugf("Withdraw: %s, pending balance is %s", this.excel.ContractAddress, utils.ToStringByPrecise(balance, erc20.Decimals))
		return this.newWithdrawErc20Tx(erc20.ContractAddr, to, withdrawAmount, DEFAULT_GAS_PRICE)
	}
}

func (this *EthManager) estimateGasLimit(tokenType string, contractAddr, to ethComm.Address, amount *big.Int, gasPrice *big.Int) (uint64, error) {
	if tokenType == config.ETH {
		callMsg := ethereum.CallMsg{
			From: this.account.Address, To: &to, Gas: 0, GasPrice: gasPrice,
			Value: amount, Data: []byte{},
		}
		gasLimit, err := this.ethClient.EstimateGas(context.Background(), callMsg)
		if err != nil {
			return 0, fmt.Errorf("newWithdrawEthTx: pre-execute failed, err: %s", err)
		}
		return gasLimit * 2, nil
	} else if tokenType == config.ERC20 {
		txData, err := this.Erc20Abi.Pack("transfer", to, amount)
		if err != nil {
			return 0, fmt.Errorf("newWithdrawErc20Tx: pack tx data failed, err: %s", err)
		}
		to := ethComm.HexToAddress("0xd46e8dd67c5d32be8058bb8eb970870f07244567")
		callMsg := ethereum.CallMsg{
			From: this.account.Address, To: &to, Gas: 0, GasPrice: gasPrice,
			Value: big.NewInt(0), Data: txData,
		}
		gasLimit, err := this.ethClient.EstimateGas(context.Background(), callMsg)
		if err != nil {
			return 0, fmt.Errorf("newWithdrawErc20Tx: pre-execute failed, err: %s", err)
		}
		return gasLimit * 2, nil
	} else {
		return 0, fmt.Errorf("unknown token type: %s", tokenType)
	}
}

func (this *EthManager) newWithdrawErc20Tx(contractAddr, to ethComm.Address, amount *big.Int, gasPrice *big.Int) (string, []byte, error) {
	txData, err := this.Erc20Abi.Pack("transfer", to, amount)
	if err != nil {
		return "", nil, fmt.Errorf("newWithdrawErc20Tx: pack tx data failed, err: %s", err)
	}
	gasLimit, err := this.estimateGasLimit(config.ERC20, contractAddr, to, amount, gasPrice)
	if err != nil {
		return "", nil, fmt.Errorf("EstimateGasLimit error:%s", err)
	}
	gasLimit = gasLimit * 2
	gasL := new(big.Int).SetUint64(gasLimit)
	fee := new(big.Int).Mul(gasL, DEFAULT_GAS_PRICE)
	log.Infof("fee:%s", utils.ToStringByPrecise(fee, config.ETH_DECIMALS))
	return this.newTx(contractAddr, big.NewInt(0), gasLimit, gasPrice, txData)
}

func (this *EthManager) newWithdrawEthTx(to ethComm.Address, amount *big.Int, gasPrice *big.Int) (string, []byte, error) {
	gasLimit, err := this.estimateGasLimit(config.ETH, to, to, amount, gasPrice)
	if err != nil {
		return "", nil, fmt.Errorf("EstimateGasLimit error:%s", err)
	}
	return this.newTx(to, amount, gasLimit, gasPrice, []byte{})
}

func (this *EthManager) newTx(to ethComm.Address, value *big.Int, gasLimit uint64, gasPrice *big.Int,
	txData []byte) (string, []byte, error) {
	err := this.keyStore.TimedUnlock(this.account, config.PASSWORD, time.Minute)
	if err != nil {
		return "", nil, fmt.Errorf("newTx: unlock account failed, err: %s", err)
	}
	tx := types.NewTransaction(this.nonce, to, value, gasLimit, gasPrice, txData)
	signedTx, err := this.keyStore.SignTx(this.account, tx, nil)
	if err != nil {
		return "", nil, fmt.Errorf("newTx: sign tx failed, err: %s", err)
	}
	log.Debugf("newTx: account %s, hash %s, nonce %d", this.account.Address.String(), signedTx.Hash().String(),
		this.nonce)
	this.nonce++
	txBuffer := new(bytes.Buffer)
	err = signedTx.EncodeRLP(txBuffer)
	if err != nil {
		return "", nil, fmt.Errorf("EncodeRLP error: %s", err)
	}
	return signedTx.Hash().String(), txBuffer.Bytes(), nil
}

type rpcTransaction struct {
	tx *types.Transaction
	txExtraInfo
}

type txExtraInfo struct {
	BlockNumber *string         `json:"blockNumber,omitempty"`
	BlockHash   *common.Hash    `json:"blockHash,omitempty"`
	From        *common.Address `json:"from,omitempty"`
}

func (this *EthManager) GetTxTime(txHash string) (uint32, error) {
	var jsonRes *rpcTransaction
	err := this.txTimeClient.CallContext(context.Background(), &jsonRes, "eth_getTransactionByHash", txHash)
	if err != nil {
		return 0, err
	} else if jsonRes == nil {
		return 0, fmt.Errorf("jsonRes is nil")
	}
	block, err := this.ethClient.BlockByHash(context.Background(), *jsonRes.BlockHash)
	if err != nil {
		return 0, err
	}
	return uint32(block.Time()), nil
}

type EthRes struct {
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Result  []*TxRes `json:"result"`
}

type TxRes struct {
	BlockNumber string `json:"blockNumber"`
	TimeStamp   string `json:timeStamp`
}

func (this *EthManager) SendTx(txHex []byte) (string, error) {
	tx := &types.Transaction{}
	err := rlp.DecodeBytes(txHex, tx)
	if err != nil {
		return "", fmt.Errorf("Decode error: %s", err)
	}
	err = this.ethClient.SendTransaction(context.Background(), tx)
	if err != nil {
		return "", fmt.Errorf("[SendTx] send tx %s failed, err: %s", tx.Hash().String(), err)
	}
	return tx.Hash().String(), nil
}

func (this *EthManager) reSendTx(txHash string) {
	txInfo, err := this.db.QueryTxHexByTxHash(txHash)
	if err != nil {
		log.Errorf("[reSendTx] QueryTxHexByTxHash error:%s", err)
		return
	}
	if txInfo.TxHex != "" {
		txHex, _ := common3.HexToBytes(txInfo.TxHex)
		_, err := this.SendTx(txHex)
		if err != nil {
			log.Errorf("[reSendTx] error:%s, txHash: %s", err, txHash)
			return
		}
	}
}

func (this *EthManager) VerifyTx(txHash string, retryLimit int) (bool, error) {
	hash := common.HexToHash(txHash)
	retry := 0
	pendingAmt := 0
	for {
		_, isPending, err := this.ethClient.TransactionByHash(context.Background(), hash)
		if err == ethereum.NotFound {
			log.Errorf("[VerifyTx] TransactionByHash error: %s, txHash:%s", err, txHash)
			return false, ethereum.NotFound
		}
		if err != nil {
			if retry >= retryLimit {
				return false, err
			}
			log.Errorf("[VerifyTx] error: %s, txHash: %s, retry: %d", err, txHash, retry)
			retry++
			time.Sleep(time.Second * config.EthSleepTime)
			continue
		}
		if isPending {
			if pendingAmt < config.PendingLimit {
				pendingAmt++
				log.Infof("[VerifyTx] TransactionByHash error:%s, txHash:%s, isPending:%t", err, txHash, isPending)
			} else {
				this.reSendTx(txHash)
				return this.VerifyTx(txHash, config.RetryLimit)
			}
			time.Sleep(time.Duration(config.EthSleepTime) * time.Second)
			continue
		}
		receipt, err := this.ethClient.TransactionReceipt(context.Background(), hash)
		if err == ethereum.NotFound {
			log.Errorf("[VerifyTx]TransactionReceipt error: %s, txHash:%s", err, txHash)
			return false, ethereum.NotFound
		}
		if err != nil {
			log.Errorf("[VerifyTx]TransactionReceipt error: %s, txHash: %s", err, txHash)
			return false, err
		}
		if receipt != nil && receipt.Status == types.ReceiptStatusSuccessful {
			return true, nil
		} else {
			if receipt != nil {
				bs, err := receipt.MarshalJSON()
				if err != nil {
					log.Errorf("[VerifyTx]MarshalJSON error: %s", err)
				} else {
					log.Errorf("[VerifyTx]verify tx failed, err: %s", string(bs))
				}
			}
			return false, fmt.Errorf("tx failed")
		}
	}
}
