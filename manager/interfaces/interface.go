package interfaces

import "github.com/ontio/bonus/common"

type WithdrawManager interface {
	NewWithdrawTx(destAddr string, amount string) (string, []byte, error)
	SendTx(txHex []byte) (string, error)
	VerifyTx(txHash string) bool
	SetContractAddress(address string) error
	StartTransfer()
	StartHandleTxTask()
	GetAdminAddress() string
	EstimateFee() (string, error)
	GetTxTime(txHash string) (uint32, error)
	GetAdminBalance() (map[string]string, error)
	WithdrawToken() error
	ComputeSum() (string, error)
	GetStatus() common.TransferStatus
	VerifyAddress(address string) bool
}
