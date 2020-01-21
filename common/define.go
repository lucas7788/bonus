package common

type TransferParam struct {
	Address string
	Amount  string
}

type TransactionInfo struct {
	EventType       string
	TokenType       string
	ContractAddress string
	Address         string
	Amount          string
	TxHash          string
	TxTime          uint32
	TxHex           string
	TxResult        TxResult
	ErrorDetail     string
}

type GetDataByEventType struct {
	TxInfo          []*TransactionInfo
	Admin           string
	EstimateFee     string
	Sum             string
	AdminBalance    string
	EventType       string
	TokenType       string
	ContractAddress string
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

type ExcelParam struct {
	BillList        []*TransferParam
	TokenType       string
	ContractAddress string
	EventType       string
	Admin           string
	EstimateFee     string
	Sum             string
	AdminBalance    string
}
