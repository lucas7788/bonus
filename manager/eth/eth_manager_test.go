package eth

import (
	"testing"

	"fmt"
	"github.com/ontio/bonus/common"
	"github.com/ontio/bonus/config"
	"github.com/stretchr/testify/assert"
	"math/big"
	"time"
)

func TestEthManager_NewWithdrawTx(t *testing.T) {

	gwei := new(big.Int).SetUint64(uint64(1000000000))
	fmt.Println(gwei)
	a, _ := new(big.Int).SetString("1000000000", 10)
	fmt.Println(a)

	res := new(big.Int)
	res = res.Div(DEFAULT_GAS_PRICE, gwei)
	fmt.Println(res)
	eth := &config.Eth{
		KeyStore:       "../../wallets/eth",
		RpcAddrMainNet: "http://onto-eth.ont.io:10331",
		RpcAddrTestNet: "https://ropsten.infura.io/v3/3425c463d2f1455c8c260b990c71a888",
	}

	eatp := &common.ExcelParam{
		BillList:  nil,
		TokenType: config.ETH,
		EventType: "test",
	}
	manager, err := NewEthManager(eth, eatp, config.TestNet)
	if err != nil {
		fmt.Println("NewEthManager err:", err)
		return
	}
	fmt.Println(manager.account.Address.String())
	fmt.Println(manager.GetAdminBalance())
	err = manager.WithdrawToken("0x4e7946D1Ee8f8703E24C6F3fBf032AD4459c4648", config.ETH)
	assert.Nil(t, err)

}

func TestEthManager_Withdraw(t *testing.T) {
	//c := &config.EthToken{
	//	//TokenName:     config.ERC20,
	//	TokenName:"ENU",
	//	ContractAddr:"0x275b69aa7c8c1d648a0557656bce1c286e69a29d",
	//	//ContractAddr: "0x247f83Ade8379A5bf4c98c18D68E64Cdf08E7CD9",
	//	}
	eth := &config.Eth{
		KeyStore:       "../../wallets/eth",
		RpcAddrMainNet: "http://onto-eth.ont.io:10331",
		RpcAddrTestNet: "https://ropsten.infura.io/v3/3425c463d2f1455c8c260b990c71a888",
	}

	manager, err := NewEthManager(eth, nil, config.MainNet)
	if err != nil {
		fmt.Println("NewEthManager err:", err)
		return
	}
	return
	txhash, txHex, err := manager.NewWithdrawTx("0x1F8aD8DDC9b248f46C34F66a28b14c3B867f02e3", "1.0", config.ETH)
	if err != nil {
		fmt.Println("[NewWithdrawTx] err:", err)
		return
	}
	hash, err := manager.SendTx(txHex)
	if err != nil {
		fmt.Println("[SendTx] err:", err)
		return
	}
	fmt.Println("start verify tx:", time.Now().Second())
	boo := manager.VerifyTx(hash, 6)
	if boo {
		fmt.Println("tx success txhash:", txhash)
	} else {
		fmt.Println("tx failed")
	}
	fmt.Println("end verify tx:", time.Now().Second())
}
func TestNewEthManager(t *testing.T) {
	//c := &config.EthToken{
	//	TokenName:    config.ERC20,
	//	ContractAddr: "0x247f83Ade8379A5bf4c98c18D68E64Cdf08E7CD9"}
	eth := &config.Eth{
		KeyStore: "./testdata2/wallets/eth",
		//Account:  "0x79dd7951f80c7184259935272e2fe69fa00f2aae",
		RpcAddrTestNet: "https://ropsten.infura.io/v3/3425c463d2f1455c8c260b990c71a888",
	}
	manager, err := NewEthManager(eth, nil, config.TestNet)
	assert.Nil(t, err)
	assert.NotEqual(t, nil, manager)
}

func TestEthManager_GetTxTime(t *testing.T) {
	//c := &config.EthToken{
	//	TokenName:    config.ERC20,
	//	ContractAddr: "0x247f83Ade8379A5bf4c98c18D68E64Cdf08E7CD9"}
	eth := &config.Eth{
		KeyStore: "./testdata2/wallets/eth",
		//Account:  "0x79dd7951f80c7184259935272e2fe69fa00f2aae",
		RpcAddrTestNet: "https://ropsten.infura.io/v3/3425c463d2f1455c8c260b990c71a888",
	}
	manager, _ := NewEthManager(eth, nil, config.TestNet)

	ti, err := manager.GetTxTime("0x4df8e59e05a1f89cfa70b0db8d00c70e623cccbea07b53c36bc5b5ac041ca4f8")
	assert.Nil(t, err)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	assert.NotEqual(t, ti, 0)
	fmt.Println("ti:", ti)
}

func TestEthManager_EstimateFee(t *testing.T) {
	eth := &config.Eth{
		KeyStore: "./testdata2/wallets/eth",
		//Account:  "0x79dd7951f80c7184259935272e2fe69fa00f2aae",
		//RpcAddr: "https://ropsten.infura.io/v3/3425c463d2f1455c8c260b990c71a888",
		RpcAddrTestNet: "http://onto-eth.ont.io:10331",
	}
	tp := &common.TransferParam{
		Id:      1,
		Address: "0x95FB49AE2DEC0D2a37b27033742fd99915faF6A1",
		Amount:  "100",
	}
	tps := make([]*common.TransferParam, 0)
	tps = append(tps, tp)
	eatp := &common.ExcelParam{
		BillList:        tps,
		TokenType:       config.ERC20,
		ContractAddress: "0xdac17f958d2ee523a2206206994597c13d831ec7",
		EventType:       "sss",
		Admin:           "",
		EstimateFee:     "",
		Sum:             "",
		AdminBalance:    nil,
	}
	manager, err := NewEthManager(eth, eatp, config.TestNet)
	if err != nil {
		fmt.Println("NewEthManager:", err)
	}
	fee, err := manager.EstimateFee("", 1)
	if err != nil {
		fmt.Println("EstimateFee:", err)
	}
	fmt.Println("fee:", fee)
}
