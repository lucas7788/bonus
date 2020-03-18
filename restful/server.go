package restful

import (
	"github.com/qiangxue/fasthttp-routing"
)

const (
	POST_EXECL              = "/api/v1/uploadexecl"
	POST_TRANSFER           = "/api/v1/transfer"
	POST_DATA_BY_EVENT_TYPE = "/api/v1/getdatabyeventtype"
	POST_WITHDRAW           = "/api/v1/withdraw"
	POST_SET_GAS_PRICE      = "/api/v1/setgasprice"
	GET_ALL_EVENT_TYPE      = "/api/v1/getalleventtype/<nettype>"
	GET_GAS_PRICE           = "/api/v1/getgasprice"
	Get_Excel_Param         = "/api/v1/getexcelparam"
	Get_Tansfer_Progress    = "/api/v1/gettransferprogress"
	GET_ADMIN_BALANCE_BY_EVENT_TYPE = "/api/v1/getadminbalancebyeventtype/<eventtype>/<nettype>"
)

//init restful server
func InitRouter() *routing.Router {
	router := routing.New()
	router.Post(POST_EXECL, UpLoadExcel)
	router.Post(POST_TRANSFER, Transfer)
	router.Post(POST_DATA_BY_EVENT_TYPE, GetTxInfoByEventType)
	router.Post(POST_WITHDRAW, Withdraw)
	router.Post(POST_SET_GAS_PRICE, SetGasPrice)

	router.Get(GET_ALL_EVENT_TYPE, GetAllEventType)
	router.Get(GET_GAS_PRICE, GetGasPrice)
	router.Get(Get_Excel_Param, GetExcelParamByEvtType)
	router.Get(Get_Tansfer_Progress, GetTransferProgress)
	router.Get(GET_ADMIN_BALANCE_BY_EVENT_TYPE, GetAdminBalanceByEventType)
	return router
}
