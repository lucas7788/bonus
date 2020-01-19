package eth

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/CandyDrop/utils"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ontio/ontology/common/log"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"testing"
	"time"
)

const (
	apiKey     = "INEP15BD1NM8P5S5T7TQHNYN513I2M7S9A"
	address    = "0x09ee99863344f4ff60127691db5eca33c4b7889b"
	startBlock = 7667141
	endBlock   = 7667144

	receipt_failed  = "0"
	receipt_success = "1"
)

type CommonResp struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Result  []*RpcTransaction `json:"result"`
}

type RpcTransaction struct {
	BlockNumber       string `json:"blockNumber"`
	BlockHash         string `json:"blockHash"`
	TimeStamp         string `json:"timeStamp"`
	Hash              string `json:"hash"`
	Nonce             string `json:"nonce"`
	TransactionIndex  string `json:"transactionIndex"`
	From              string `json:"from"`
	To                string `json:"to"`
	Value             string `json:"value"`
	Gas               string `json:"gas"`
	GasPrice          string `json:"gasPrice"`
	Input             string `json:"input"`
	ContractAddress   string `json:"contractAddress"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	TxreceiptStatus   string `json:"txreceipt_status"`
	GasUsed           string `json:"gasUsed"`
	Confirmations     string `json:"confirmations"`
	IsError           string `json:"isError"`
}

type FailedTx struct {
	BlockNumber string
	Hash        string
	To          string
	Amount      string
}

func TestFetchFailedTx(t *testing.T) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   5,
			DisableKeepAlives:     false,
			IdleConnTimeout:       time.Second * 300,
			ResponseHeaderTimeout: time.Second * 300,
		},
		Timeout: time.Second * 300,
	}
	url := fmt.Sprintf("http://api.etherscan.io/api?module=account&action=txlist&"+
		"address=%s&startblock=%d&endblock=%d&sort=asc&apikey=%s", address, startBlock, endBlock, apiKey)
	resp, err := httpClient.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	t.Logf(string(data))
	commonResp := &CommonResp{}
	err = json.Unmarshal(data, commonResp)
	if err != nil {
		t.Fatal(err)
	}
	erc20Abi, err := abi.JSON(strings.NewReader(Erc20ABI))
	if err != nil {
		t.Fatal(err)
	}
	method, ok := erc20Abi.Methods["transfer"]
	if !ok {
		t.Fatal("abi not contain transfer")
	}
	wholeFailedTx := make([]*FailedTx, 0)
	count := 0
	wholeFailedAmount := big.NewInt(0)
	for _, rpcTx := range commonResp.Result {
		if rpcTx.TxreceiptStatus == receipt_failed {
			inputData, err := hex.DecodeString(rpcTx.Input[2:])
			if err != nil {
				t.Fatal(err)
			}
			input, err := method.Inputs.UnpackValues(inputData[4:])
			if err != nil {
				t.Fatal(err)
			}
			if len(input) != 2 {
				t.Fatal("input error")
			}
			to, ok := input[0].(common.Address)
			if !ok {
				t.Fatal("decode to param failed")
			}
			amount, ok := input[1].(*big.Int)
			if !ok {
				t.Fatal("decode amount param failed")
			}
			amountString := utils.ToStringByPrecise(amount, 18)
			t.Logf("unpack to %s, amount %s", to.String(), amountString)
			failedTx := &FailedTx{
				BlockNumber: rpcTx.BlockNumber,
				Hash:        rpcTx.Hash,
				To:          to.String(),
				Amount:      amountString,
			}
			wholeFailedTx = append(wholeFailedTx, failedTx)
			count++
			wholeFailedAmount.Add(wholeFailedAmount, amount)
		}
	}
	fileData, _ := json.Marshal(wholeFailedTx)
	err = ioutil.WriteFile("failed_tx.json", fileData, 0666)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("failed people %d, whole num %s", count, utils.ToStringByPrecise(wholeFailedAmount, 18))
}

func TestAbi(t *testing.T) {
	erc20Abi, err := abi.JSON(strings.NewReader(Erc20ABI))
	if err != nil {
		t.Fatal(err)
	}
	to := common.HexToAddress("0x09ee99863344f4ff60127691db5eca33c4b7889b")
	amount := utils.ToIntByPrecise("1000", 18)
	txData, err := erc20Abi.Pack("transfer", to, amount)
	t.Logf("txData is %x", txData)
	method := erc20Abi.Methods["transfer"]
	input, err := method.Inputs.UnpackValues(txData[4:])
	if err != nil {
		t.Fatal(err)
	}
	if len(input) != 2 {
		t.Fatal("input error")
	}
	t.Logf("unpack to %s, amount %s", input[0].(common.Address).String(),
		utils.ToStringByPrecise(input[1].(*big.Int), 18))
}

func TestPendingWithdraw(t *testing.T) {
	client, err := ethclient.Dial("https://mainnet.infura.io/v3/3425c463d2f1455c8c260b990c71a888")
	if err != nil {
		t.Fatal(err)
	}
	contractAddr := common.HexToAddress("0xdcD85914b8aE28c1E62f1C488E1D968D5aaFfE2b")
	contract, err := NewErc20(contractAddr, client)
	if err != nil {
		t.Fatal(err)
	}
	account := common.HexToAddress("0x09ee99863344f4ff60127691db5eca33c4b7889b")
	callOpt := &bind.CallOpts{Pending: true}
	topDecimals, err := contract.Decimals(callOpt)
	if err != nil {
		t.Fatal(err)
	}
	for {
		balance, err := contract.BalanceOf(callOpt, account)
		if err != nil {
			t.Fatal(err)
		}
		log.Infof("top pending balance is %s", utils.ToStringByPrecise(balance, topDecimals.Uint64()))
		<-time.After(1 * time.Second)
	}
}
