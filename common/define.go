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

type TransferStatus byte

const (
	NotTransfer TransferStatus = iota
	Transfering
	Transfered
)

type TxResult byte

const (
	NotBuild TxResult = iota //
	BuildTxFailed
	SendFailed
	OneTransfering
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
