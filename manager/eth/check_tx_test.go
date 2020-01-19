package eth

import (
	"encoding/json"
	"github.com/CandyDrop/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ontio/ontology/common/log"
	"io/ioutil"
	"math/big"
	"testing"
)

func TestCheckTopTx(t *testing.T) {
	filePath := "../../candy_usertokenlog2019051818.txt"
	withdrawTxs, err := ParseSqlData(filePath)
	if err != nil {
		t.Fatal(err)
	}
	client, err := ethclient.Dial("https://mainnet.infura.io/v3/3425c463d2f1455c8c260b990c71a888")
	if err != nil {
		t.Fatal(err)
	}
	erc20Contract, err := NewErc20(common.HexToAddress("0xdcd85914b8ae28c1e62f1c488e1d968d5aaffe2b"), client)
	if err != nil {
		t.Fatal(err)
	}
	decimals, err := erc20Contract.Decimals(&bind.CallOpts{})
	if err != nil {
		t.Fatal(err)
	}
	failedTxs := make([]*FailedTx, 0)
	wholeFailedAmount := new(big.Int)
	for _, withdraw := range withdrawTxs {
		success, err := CheckErc20WithdrawTx(withdraw, client, erc20Contract, decimals.Uint64())
		if !success {
			if err != nil {
				log.Errorf("Withdraw %s err: %s", withdraw.String(), err)
			} else {
				log.Warnf("Withdraw %s failed", withdraw.String())
			}
			amount := utils.ToIntByPrecise(withdraw.Amount, decimals.Uint64())
			wholeFailedAmount.Add(wholeFailedAmount, amount)
			failedTxs = append(failedTxs, &FailedTx{Hash: withdraw.Hash, To: withdraw.To,
				Amount: utils.ToStringByPrecise(amount, decimals.Uint64())})
		}
	}
	log.Infof("whole failed tx num %d, amount %s", len(failedTxs),
		utils.ToStringByPrecise(wholeFailedAmount, decimals.Uint64()))
	failedTxData, err := json.Marshal(failedTxs)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile("failed_tx.json", failedTxData, 0666)
	if err != nil {
		t.Fatal(err)
	}
}
