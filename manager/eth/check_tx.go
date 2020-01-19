package eth

import (
	"context"
	"fmt"
	"github.com/CandyDrop/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"io/ioutil"
	"strings"
)

type WithdrawTx struct {
	LogId  string
	Amount string
	To     string
	Hash   string
}

func (this *WithdrawTx) String() string {
	return fmt.Sprintf("LogId %s, Amount %s, To %s, Hash %s", this.LogId, this.Amount, this.To, this.Hash)
}

func ParseSqlData(filePath string) ([]*WithdrawTx, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("ParseSqlData: err %s", err)
	}
	fileContent := string(data)
	lineData := strings.Split(fileContent, "\n")
	result := make([]*WithdrawTx, 0)
	for i := 1; i < len(lineData); i++ {
		withdrawData := strings.Split(lineData[i], "	")
		if len(withdrawData) != 10 {
			return nil, fmt.Errorf("ParseSqlData: data format unsupport")
		}
		result = append(result, &WithdrawTx{
			LogId:  withdrawData[0],
			Amount: withdrawData[4],
			To:     withdrawData[5],
			Hash:   withdrawData[9],
		})
	}
	return result, nil
}

func CheckErc20WithdrawTx(withdraw *WithdrawTx, client *ethclient.Client, erc20 *Erc20, decimals uint64) (bool, error) {
	hash := common.HexToHash(withdraw.Hash)
	txRec, err := client.TransactionReceipt(context.Background(), hash)
	if err == ethereum.NotFound {
		return false, err
	}
	if err != nil {
		return false, fmt.Errorf("CheckErc20WithdrawTx: cannot get withdraw %s receipt, err %s", withdraw.String(), err)
	}
	if txRec.Status == types.ReceiptStatusFailed {
		return false, nil
	}
	isMatch := false
	toAddr := common.HexToAddress(withdraw.To)
	amount := utils.ToIntByPrecise(withdraw.Amount, decimals)
	for _, txLog := range txRec.Logs {
		transferEvent := &Erc20Transfer{}
		err = erc20.Erc20Caller.contract.UnpackLog(transferEvent, "Transfer", *txLog)
		if err == nil {
			if transferEvent.To == toAddr && transferEvent.Value.Cmp(amount) == 0 {
				isMatch = true
				break
			}
		}
	}
	return isMatch, nil
}
