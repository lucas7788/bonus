package restful

import (
	"github.com/qiangxue/fasthttp-routing"
)

const (
	POST_EXECL                      = "/api/v1/uploadexecl"
	POST_TRANSFER                   = "/api/v1/transfer"
	POST_DATA_BY_EVENT_TYPE         = "/api/v1/getdatabyeventtype"
	GET_ALL_EVENT_TYPE              = "/api/v1/getalleventtype"
	GET_ADMIN_BALANCE_BY_EVENT_TYPE = "/api/v1/getadminbalancebyeventtype/<eventtype>"
)

//init restful server
func InitRouter() *routing.Router {
	router := routing.New()

	router.Post(POST_EXECL, UpLoadExcel)
	router.Post(POST_TRANSFER, Transfer)
	router.Post(POST_DATA_BY_EVENT_TYPE, GetDataByEventType)
	router.Get(GET_ADMIN_BALANCE_BY_EVENT_TYPE, GetAdminBalanceByEventType)
	router.Get(GET_ALL_EVENT_TYPE, GetAllEventType)
	return router
}
