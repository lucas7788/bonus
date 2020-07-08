package interfaces

import (
	"github.com/ontio/bonus/bonus_db"
	"github.com/ontio/bonus/common"
	"math/big"
)

type WithdrawManager interface {
	NewWithdrawTx(destAddr string, amount *big.Int, tokenType string) (string, []byte, error)
	SendTx(txHex []byte) (string, error)
	VerifyTx(txHash string, retryLimit int) (bool, error)
	StartTransfer()
	GetAdminAddress() string
	EstimateFee(tokenType string, total int) (string, error)
	GetTxTime(txHash string) (uint32, error)
	GetAdminBalance() (map[string]string, error)
	WithdrawToken(address string, tokenType string) error
	ComputeSum() (string, error)
	GetStatus() common.TransferStatus
	GetWithdrawStatus() int
	VerifyAddress(address string) bool
	GetNetType() string
	GetTotal() int
	GetExcelParam() *common.ExcelParam
	Store() error
	Stop()
	CloseDB()
	QueryTransferProgress() (map[string]int, error)
	QueryTxInfo(start, end int, txResult common.TxResult) ([]*common.TransactionInfo, int, error)
	SetGasPrice(gasPrice uint64) error
	GetGasPrice() uint64
	SetDB(db *bonus_db.BonusDB)
}
