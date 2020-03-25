package common

type TransferParam struct {
	Id      int    `json:"id"`
	Address string `json:"address"`
	Amount  string `json:"amount"`
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
	Total           int
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
	AllStatus
)

type ExcelParam struct {
	BillList        []*TransferParam `json:"bill_list"`
	TokenType       string           `json:"token_type"`
	ContractAddress string           `json:"contract_address"`
	EventType       string           `json:"event_type"`
	Admin           string           `json:"admin"`
	EstimateFee     string           `json:"estimate_fee"`
	NetType         string           `json:"net_type"`
	Sum             string
	AdminBalance    map[string]string
	Total           int
}

type WithdrawParam struct {
	EventType string
	Address   string
	TokenType string
	NetType   string
}
