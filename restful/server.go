package restful

import (
	"github.com/qiangxue/fasthttp-routing"
)

const (
	POST_EXECL      = "/api/v1/uploadexecl"
	POST_TRANSFER   = "/api/v1/start"
	GET_ALL_KEY     = "/api/v1/getallkey"
	GET_DATA_BY_KEY = "/api/v1/getdatabykey"
)

//init restful server
func InitRouter() *routing.Router {
	router := routing.New()
	router.Post(POST_EXECL, UpLoadExcelAndTransfer)
	router.Get(GET_DATA_BY_KEY+"/<key>", GetDataByKey)
	return router
}
