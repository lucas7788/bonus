package common

type TransferParam struct {
	Address string
	Amount  float64
}

type TransactionInfo struct {
	TokenType string
	Address   string
	Amount    float64
	TxHash    string
	TxHex     string
	TxResult  TxResult
}
type TxResult byte

const (
	BuildTxFailed TxResult = iota //
	NotSend
	SendFailed
	SendSuccess
	TxFailed
	TxSuccess
)


type ExcelAndTransferParam struct {
	BillList          []*TransferParam
	TokenType         string
	PrivateKey        string
	ContractAddress   string
	FileName          string
}