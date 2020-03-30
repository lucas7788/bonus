package restful

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/ontio/bonus/common"
	"github.com/ontio/bonus/config"
	"github.com/ontio/bonus/manager"
	"github.com/ontio/bonus/manager/interfaces"
	"github.com/ontio/ontology/common/log"
	"github.com/qiangxue/fasthttp-routing"
)

var DefBonusMap = new(sync.Map) // evttype + nettype -> withdraw-mgr
// evttype -> TokenType

func UploadExcel(ctx *routing.Context) error {
	excelParam, errCode := ParseExcelParam(ctx)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}

	if _, exist := DefBonusMap.Load(excelParam.EventType); exist {
		return writeResponse(ctx, ResponsePack(DuplicateEventType))
	}

	excelParam.ResetTransferListID()

	mgr, err := manager.CreateManager(excelParam, excelParam.NetType, nil)
	if err != nil {
		log.Errorf("CreateManager error: %s", err)
		return writeResponse(ctx, ResponsePack(InitManagerError))
	}

	if err := updateExcelParam(mgr, excelParam); err != SUCCESS {
		log.Errorf("updateExcelParam error: %d", err)
		return writeResponse(ctx, ResponsePack(err))
	}

	// persist mgr to event-local json file
	err = mgr.Store()
	if err != nil {
		log.Errorf("Store error: %s", err)
		return writeResponse(ctx, ResponsePack(InsertSqlError))
	}
	DefBonusMap.Store(excelParam.EventType, excelParam.TokenType)
	DefBonusMap.Store(excelParam.EventType+excelParam.NetType, mgr)
	return writeResponse(ctx, ResponseSuccess(excelParam))
}

func GetAdminBalanceByEventType(ctx *routing.Context) error {
	evtType := ctx.Param("evtty")
	netType := ctx.Param("netty")
	mgr, errCode := getTokenManager(evtType, netType)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	adminBalance, err := mgr.GetAdminBalance()
	if err != nil {
		log.Errorf("GetAdminBalance error: %s", err)
		return writeResponse(ctx, ResponsePack(GetAdminBalanceError))
	}
	return writeResponse(ctx, ResponseSuccess(adminBalance))
}

func GetGasPrice(ctx *routing.Context) error {
	netty := ctx.Param("netty")
	evtty := ctx.Param("evtty")
	mgr, errCode := getTokenManager(evtty, netty)

	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	return writeResponse(ctx, ResponseSuccess(mgr.GetGasPrice()))
}

func SetGasPrice(ctx *routing.Context) error {
	gasPriceInt, evtTy, netTy, errCode := ParseSetGasPriceParam(ctx)
	if errCode != SUCCESS {
		log.Errorf("ParseSetGasPriceParam error ")
		return writeResponse(ctx, ResponsePack(errCode))
	}
	mgr, errCode := getTokenManager(evtTy, netTy)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	if err := mgr.SetGasPrice(uint64(gasPriceInt)); err != nil {
		log.Errorf("set gas price failed: %s", err)
		return writeResponse(ctx, ResponsePack(SetGasPriceFailed))
	}
	return writeResponse(ctx, ResponsePack(SUCCESS))
}

func Transfer(ctx *routing.Context) error {
	eventType, netType, errCode := ParseTransferParam(ctx)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	mgr, errCode := getTokenManager(eventType, netType)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	log.Info("transfer status:", mgr.GetStatus())
	if mgr.GetStatus() == common.Transfering {
		return writeResponse(ctx, ResponsePack(Transfering))
	}
	mgr.StartTransfer()
	log.Info("start transfer success")
	return writeResponse(ctx, ResponsePack(SUCCESS))
}

func Withdraw(ctx *routing.Context) error {
	withdrawParam, errCode := ParseWithdrawParam(ctx)
	if errCode != SUCCESS {
		log.Errorf("[Withdraw] ParseWithdrawParam error:%s", errCode)
		return writeResponse(ctx, ResponsePack(errCode))
	}

	mgr, errCode := getTokenManager(withdrawParam.EventType, withdrawParam.NetType)
	if errCode != SUCCESS {
		log.Errorf("[Withdraw] get token mgr %s/%s error:%d", withdrawParam.EventType, withdrawParam.NetType, errCode)
		return writeResponse(ctx, ResponsePack(errCode))
	}

	if mgr.VerifyAddress(withdrawParam.Address) == false {
		return writeResponse(ctx, ResponsePack(AddressIsWrong))
	}

	if mgr.GetStatus() == common.Transfering {
		return writeResponse(ctx, ResponsePack(Transfering))
	}
	err := mgr.WithdrawToken(withdrawParam.Address, strings.ToUpper(withdrawParam.TokenType))
	if err != nil {
		log.Errorf("WithdrawToken failed, error: %s", err)
		res := ResponsePack(WithdrawTokenFailed)
		res["Result"] = err
		return writeResponse(ctx, res)
	}
	return writeResponse(ctx, ResponsePack(SUCCESS))
}

func GetEventType(ctx *routing.Context) error {
	eventdirs, err := config.GetAllEventDirs()
	if err != nil {
		return writeResponse(ctx, QueryExcelParamByEventType)
	}

	events := make([]string, 0)
	for _, eventName := range eventdirs {
		e := ""
		for _, tokenName := range config.SupportedTokenTypes {
			if strings.HasPrefix(eventName, tokenName+"_") {
				e = eventName[len(tokenName)+1:]
				break
			}
		}
		if e != "" {
			events = append(events, e)
		}
	}

	return writeResponse(ctx, ResponseSuccess(events))
}

func GetTransferProgress(ctx *routing.Context) error {
	evtty := ctx.Param("evtty")
	netty := ctx.Param("netty")
	mgr, errCode := getTokenManager(evtty, netty)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	res, err := mgr.QueryTransferProgress()
	if err != nil {
		log.Errorf("[GetTransferProgress] QueryTransferProgress failed: %s", err)
		return writeResponse(ctx, ResponsePack(QueryTransferProgressFailed))
	}
	res["evtStatus"] = int(mgr.GetStatus())
	return writeResponse(ctx, ResponseSuccess(res))
}

func GetExcelParamByEvtType(ctx *routing.Context) error {
	param, errCode := ParseQueryExcelParam(ctx)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	mgr, errCode := getTokenManager(param.EvtType, param.NetType)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}

	excelParam := mgr.GetExcelParam()
	param2 := &common.ExcelParam{
		BillList:        excelParam.BillList,
		TokenType:       excelParam.TokenType,
		ContractAddress: excelParam.ContractAddress,
		EventType:       excelParam.EventType,
		NetType:         param.NetType,
	}

	if err := updateExcelParam(mgr, param2); err != SUCCESS {
		log.Errorf("[GetExcelParamByEvtType]updateExcelParam error: %d", err)
		return writeResponse(ctx, ResponsePack(err))
	}
	if param.PageSize != 0 {
		if param.PageNum <= 0 {
			param.PageNum = 1
		}
		start := (param.PageNum - 1) * param.PageSize
		end := start + param.PageSize
		if start > len(excelParam.BillList) {
			param2.BillList = nil
		} else if end > len(excelParam.BillList) {
			param2.BillList = excelParam.BillList[start:]
		} else {
			param2.BillList = excelParam.BillList[start:end]
		}
	}
	return writeResponse(ctx, ResponseSuccess(param2))
}

func GetTxInfoByEventType(ctx *routing.Context) error {
	param, errCode := ParseQueryTxInfoParam(ctx)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	mgr, errCode := getTokenManager(param.EvtTy, param.NetTy)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	if param.PageNum < 1 {
		param.PageNum = 1
	}
	start := (param.PageNum - 1) * param.PageSize
	txInfo, total, err := mgr.QueryTxInfo(start, param.PageSize, param.TxResult)
	if err != nil {
		log.Errorf("QueryTxInfoByEventType error: %s", err)
		return writeResponse(ctx, ResponsePack(QueryResultByEventType))
	}

	sum, err := mgr.ComputeSum()
	if err != nil {
		log.Errorf("CreateManager error: %s", err)
		return writeResponse(ctx, ResponsePack(SumError))
	}
	ba, err := mgr.GetAdminBalance()
	if err != nil {
		log.Errorf("GetAdminBalance error: %s", err)
		return writeResponse(ctx, ResponsePack(GetAdminBalanceError))
	}
	fee, err := mgr.EstimateFee(mgr.GetExcelParam().TokenType, mgr.GetTotal())
	if err != nil {
		log.Errorf("EstimateFee error: %s", err)
		return writeResponse(ctx, ResponsePack(EstimateFeeError))
	}
	res := &common.GetTxInfoByEvtType{
		TxInfo:          txInfo,
		AdminBalance:    ba,
		Admin:           mgr.GetAdminAddress(),
		Sum:             sum,
		EstimateFee:     fee,
		EventType:       param.EvtTy,
		TokenType:       mgr.GetExcelParam().TokenType,
		ContractAddress: mgr.GetExcelParam().ContractAddress,
		NetType:         param.NetTy,
		Total:           total,
	}

	r := ResponsePack(SUCCESS)
	r["Result"] = res
	return writeResponse(ctx, r)
}

func updateExcelParam(mgr interfaces.WithdrawManager, excelParam *common.ExcelParam) int64 {
	var err error
	excelParam.Sum, err = mgr.ComputeSum()
	if err != nil {
		log.Errorf("CreateManager error: %s", err)
		return SumError
	}
	excelParam.AdminBalance, err = mgr.GetAdminBalance()
	if err != nil {
		log.Errorf("GetAdminBalance error: %s", err)
		return GetAdminBalanceError
	}
	excelParam.EstimateFee, err = mgr.EstimateFee(excelParam.TokenType, mgr.GetTotal())
	if err != nil {
		log.Errorf("EstimateFee error: %s", err)
		return EstimateFeeError
	}
	excelParam.Admin = mgr.GetAdminAddress()
	excelParam.Total = mgr.GetTotal()
	return SUCCESS
}

func loadAllHistoryEvents() error {
	eventdirs, err := config.GetAllEventDirs()
	if err != nil {
		return err
	}

	for _, eventdir := range eventdirs {
		strArr := strings.Split(eventdir, "_") // TokenType, EventType
		if _, ok := DefBonusMap.Load(strArr[1]); ok {
			// reset bouns map
			DefBonusMap = new(sync.Map)
			return fmt.Errorf("dupliate event: %s", eventdir)
		}
		DefBonusMap.Store(strArr[1], strArr[0])
	}
	return nil
}

func getTokenManager(eventType, netType string) (interfaces.WithdrawManager, int64) {

	tokenTy, ok := DefBonusMap.Load(eventType)
	if !ok || tokenTy == nil {
		return nil, NoTheEventTypeError
	}

	if mn, present := DefBonusMap.Load(eventType + netType); present && mn != nil {
		mgr, ok := mn.(interfaces.WithdrawManager)
		if !ok {
			return nil, TypeTransferError
		}
		return mgr, SUCCESS
	}
	tokenType := tokenTy.(string)
	mgr, err := manager.RecoverManager(tokenType, eventType, netType)
	if err != nil {
		log.Errorf("CreateManager error: %s", err)
		return nil, InitManagerError
	}
	DefBonusMap.Store(eventType+netType, mgr)

	return mgr, SUCCESS
}

func writeResponse(ctx *routing.Context, res interface{}) error {
	bs, err := json.Marshal(res)
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
