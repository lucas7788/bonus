package restful

import (
	"github.com/qiangxue/fasthttp-routing"
	"encoding/json"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/bonus/manager"
	"github.com/ontio/bonus/db"
)

func UpLoadExcelAndTransfer(ctx *routing.Context) error {
	arg, errCode := ParseExcelAndTransferParam(ctx)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	log.Info("args:", arg)
	mgr,err := manager.InitManager(arg)
	if err != nil {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	err = db.InsertSql(arg.BillList, arg.FileName)
	if err != nil {
		log.Errorf("InsertSql error: %s", err)
		return err
	}
	mgr.StartTransfer()
	res := ResponsePack(SUCCESS)
	res["Result"] = mgr.GetAdminAddress()
	return writeResponse(ctx, res)
}


func GetAllExcelFileName(ctx *routing.Context) error {
	excelFileName,err := db.QueryAllExcelFileName()
	if err != nil {
		log.Errorf("QueryAllExcelFileName error: %s", err)
		return writeResponse(ctx, ResponsePack(QueryAllExcelFileName))
	}
	res := ResponsePack(SUCCESS)
	res["Result"] = excelFileName
	return writeResponse(ctx, res)
}


func GetDataByKey(ctx *routing.Context) error {
	val := ctx.Param("excelFileName")
	res := ResponsePack(SUCCESS)
	res["Result"] = val
	return writeResponse(ctx, res)
}


func writeResponse(ctx *routing.Context, res interface{}) error {
	ctx.SetContentType("application/json;charset=utf-8")
	ctx.Set("Access-Control-Allow-Origin", "*")
	bs ,err := json.Marshal(res)
	if err != nil {
		return err
	}
	l, err := ctx.Write(bs)
	if l != len(bs) || err != nil {
		log.Errorf("write error: %s, expected length: %d, actual length: %d", err, len(bs), l)
		return err
	}
	return nil
}