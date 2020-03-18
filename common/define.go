package common

type TransferParam struct {
	Id      int
	Address string
	Amount  string
}

type TransactionInfo struct {
	Id              int
	NetType         string
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

type GetTxInfoByEvtType struct {
	TxInfo          []*TransactionInfo
	Admin           string
	EstimateFee     string
	Sum             string
	AdminBalance    map[string]string
	EventType       string
	TokenType       string
	ContractAddress string
	NetType         string
}

type TransferStatus byte

const (
	NotTransfer TransferStatus = iota
	Transfering
	Transfered
)

type TxResult byte

const (
	NotSend TxResult = iota //
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
	AdminBalance    map[string]string
	NetType         string
	Total           int
}

type WithdrawParam struct {
	EventType string
	Address   string
	TokenType string
	NetType   string
}
