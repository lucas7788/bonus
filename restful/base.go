package restful

func ResponsePack(errCode int64) map[string]interface{} {
	return map[string]interface{}{
		"Action":  "",
		"Result":  "",
		"Error":   errCode,
		"Desc":    ErrMap[errCode],
		"Version": "1.0.0",
	}
}

var ErrMap = map[int64]string{
	SUCCESS:                     "SUCCESS",
	PARA_ERROR:                  "PARAMETER ERROR",
	PARA_PARSE_ERROR:            "PARA_PARSE_ERROR",
	QueryAllEventTypeError:      "QueryAllEventTypeError",
	QueryResultByEventType:      "QueryTxInfoByEventType",
	EstimateFeeError:            "EstimateFeeError",
	InsertSqlError:              "InsertSqlError",
	TypeTransferError:           "TypeTransferError",
	NoTheEventTypeError:         "NoTheEventTypeError",
	InitManagerError:            "InitManagerError",
	SumError:                    "SumError",
	InsertEventTypeError:        "InsertEventTypeError",
	GetAdminBalanceError:        "GetAdminBalanceError",
	ExcelDuplicateAddress:       "ExcelDuplicateAddress",
	AmountIsNegative:            "AmountIsNegative",
	Transfering:                 "Transfering",
	BalanceIsNotEnough:          "BalanceIsNotEnough",
	AddressIsWrong:              "AddressIsWrong",
	WithdrawTokenFailed:         "WithdrawTokenFailed",
	QueryExcelParamByEventType:  "QueryExcelParamByEventType",
	QueryTransferProgressFailed: "QueryTransferProgressFail",
	DuplicateEventType:          "DuplicateEventType",
	NetTypeError:                "NetTypeError",
	NotSupportTokenType:         "NotSupportTokenType",
	NotExistenceEvtType:         "NotExistenceEvtType",
}

const (
	SUCCESS                     = 1
	PARA_ERROR                  = 40000
	PARA_PARSE_ERROR            = 40001
	QueryAllEventTypeError      = 40002
	QueryResultByEventType      = 40003
	EstimateFeeError            = 40004
	InsertSqlError              = 40005
	TypeTransferError           = 40006
	NoTheEventTypeError         = 40007
	InitManagerError            = 40008
	SumError                    = 40009
	InsertEventTypeError        = 40010
	GetAdminBalanceError        = 40011
	ExcelDuplicateAddress       = 40012
	AmountIsNegative            = 40013
	Transfering                 = 40014
	BalanceIsNotEnough          = 40015
	AddressIsWrong              = 40016
	WithdrawTokenFailed         = 40017
	QueryExcelParamByEventType  = 40018
	QueryTransferProgressFailed = 40019
	DuplicateEventType          = 40020
	NetTypeError                = 40021
	NotSupportTokenType         = 40022
	NotExistenceEvtType         = 40023
)
