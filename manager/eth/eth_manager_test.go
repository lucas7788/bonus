package eth

import (
	"testing"

	"fmt"
	"github.com/ontio/bonus/config"
	"github.com/stretchr/testify/assert"
	"time"
)

func TestEthManager_Withdraw(t *testing.T) {
	//c := &config.EthToken{
	//	//TokenName:     config.ERC20,
	//	TokenName:"ENU",
	//	ContractAddr:"0x275b69aa7c8c1d648a0557656bce1c286e69a29d",
	//	//ContractAddr: "0x247f83Ade8379A5bf4c98c18D68E64Cdf08E7CD9",
	//	}
	eth := &config.Eth{
		KeyStore: "../../wallets/eth",
		//RpcAddr:"http://onto-eth.ont.io:10331",
		RpcAddr: "https://ropsten.infura.io/v3/3425c463d2f1455c8c260b990c71a888",
	}

	manager, err := NewEthManager(eth, nil)
	if err != nil {
		fmt.Println("NewEthManager err:", err)
		return
	}
	txhash, txHex, err := manager.NewWithdrawTx("0x1F8aD8DDC9b248f46C34F66a28b14c3B867f02e3", "1.0")
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
	boo := manager.VerifyTx(hash)
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
		RpcAddr: "https://ropsten.infura.io/v3/3425c463d2f1455c8c260b990c71a888",
	}
	manager, err := NewEthManager(eth, nil)
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
		RpcAddr: "https://ropsten.infura.io/v3/3425c463d2f1455c8c260b990c71a888",
	}
	manager, _ := NewEthManager(eth, nil)

	ti, err := manager.GetTxTime("0x4df8e59e05a1f89cfa70b0db8d00c70e623cccbea07b53c36bc5b5ac041ca4f8")
	assert.Nil(t, err)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	assert.NotEqual(t, ti, 0)
	fmt.Println("ti:", ti)
}
