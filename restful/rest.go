package restful

import (
	"github.com/candybox-sig/log"
	"github.com/ontio/bonus/config"
	"github.com/valyala/fasthttp"
	"strconv"
)

func StartServer() {
	go func() {
		router := InitRouter()
		port := strconv.Itoa(int(config.DefConfig.RestPort))
		log.Infof("start server success, listen port: %d\n", config.DefConfig.RestPort)
		err := fasthttp.ListenAndServe(":"+port, router.HandleRequest)
		if err != nil {
			log.Errorf("ListenAndServe error: %s\n", err)
		}
	}()
}
