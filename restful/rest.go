package restful

import (
	"strconv"

	"github.com/ontio/bonus/config"
	"github.com/ontio/ontology/common/log"
	"github.com/valyala/fasthttp"
)

func StartServer() error {
	if err := loadAllHistoryEvents(); err != nil {
		return err
	}
	go func() {
		router := InitRouter()
		port := strconv.Itoa(int(config.DefConfig.RestPort))
		log.Infof("start server success, listen port: %d\n", config.DefConfig.RestPort)
		err := fasthttp.ListenAndServe(":"+port, func(ctx *fasthttp.RequestCtx) {
			ctx.Response.Header.Add("Access-Control-Allow-Headers", "Content-Type")
			ctx.Response.Header.Set("content-type", "application/json;charset=utf-8")
			ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
			ctx.Response.Header.SetContentType("application/json;charset=utf-8")
			router.HandleRequest(ctx)
		})
		if err != nil {
			log.Errorf("ListenAndServe error: %s\n", err)
		}
	}()
	return nil
}
