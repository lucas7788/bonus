package interfaces

type WithdrawManager interface {
	NewWithdrawTx(destAddr string, amount string) (string, []byte, error)
	SendTx(txHex []byte) (string, error)
	VerifyTx(txHash string) bool
	SetContractAddress(address string) error
	StartTransfer()
	StartHandleTxTask()
	GetAdminAddress() string
	GetTxTime(txHash string) (uint32, error)
}
