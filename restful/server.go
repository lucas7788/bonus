package restful

import (
	"github.com/qiangxue/fasthttp-routing"
)

const (
	POST_EXECL         = "/api/v1/uploadexecl"
	POST_TRANSFER      = "/api/v1/transfer"
	POST_WITHDRAW      = "/api/v1/withdraw"
	POST_SET_GAS_PRICE = "/api/v1/setgasprice"

	GET_EVENT_TYPE                  = "/api/v1/getevtty"
	GET_GAS_PRICE                   = "/api/v1/getgasprice/<evtty>/<netty>"
	Get_TxInfo_BY_EVENT_TYPE        = "/api/v1/gettxinfo/<evtty>/<netty>/<pagenum>/<pagesize>/<txStatus>"
	Get_Excel_Param                 = "/api/v1/getexcelparam/<evtty>/<netty>/<pagenum>/<pagesize>"
	Get_Tansfer_Progress            = "/api/v1/gettransferprogress/<evtty>/<netty>"
	GET_ADMIN_BALANCE_BY_EVENT_TYPE = "/api/v1/getadminbalance/<evtty>/<netty>"
)

//init restful server
func InitRouter() *routing.Router {
	router := routing.New()
	router.Post(POST_EXECL, UploadExcel)
	router.Post(POST_TRANSFER, Transfer)
	router.Post(POST_WITHDRAW, Withdraw)
	router.Post(POST_SET_GAS_PRICE, SetGasPrice)

	router.Get(Get_TxInfo_BY_EVENT_TYPE, GetTxInfoByEventType)
	router.Get(GET_EVENT_TYPE, GetEventType)
	router.Get(GET_GAS_PRICE, GetGasPrice)
	router.Get(Get_Excel_Param, GetExcelParamByEvtType)
	router.Get(Get_Tansfer_Progress, GetTransferProgress)
	router.Get(GET_ADMIN_BALANCE_BY_EVENT_TYPE, GetAdminBalanceByEventType)
	return router
}
