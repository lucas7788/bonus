package restful

import (
	"encoding/json"
	"github.com/ontio/bonus/common"
	"github.com/ontio/bonus/config"
	"github.com/ontio/bonus/ledger"
	"github.com/ontio/bonus/manager"
	"github.com/ontio/bonus/manager/interfaces"
	"github.com/ontio/ontology/common/log"
	"github.com/qiangxue/fasthttp-routing"
	"strings"
	"sync"
	"fmt"
)

var DefBonusMap = new(sync.Map) //projectId -> Airdrop

func UploadExcel(ctx *routing.Context) error {
	arg, netType, errCode := ParseExcelParam(ctx)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	boo := ledger.DefBonusLedger.HasExcelEvtTy(arg.EventType)
	if boo {
		log.Errorf("DuplicateEventType: %s", arg.EventType)
		return writeResponse(ctx, ResponsePack(DuplicateEventType))
	}
	ledger.DefBonusLedger.AppendExcelEvtTy(arg.EventType)
	hasInit := false
	var mgr interfaces.WithdrawManager
	if DefBonusMap == nil {
		DefBonusMap = new(sync.Map)
	} else {
		mn, ok := DefBonusMap.Load(arg.EventType + arg.NetType)
		if ok && mn != nil {
			mgr, _ = mn.(interfaces.WithdrawManager)
			hasInit = true
		}
	}
	var err error
	if !hasInit {
		mgr, err = manager.InitManager(arg, arg.NetType, nil)
		if err != nil {
			log.Errorf("InitManager error: %s", err)
			return writeResponse(ctx, ResponsePack(InitManagerError))
		}
		DefBonusMap.Store(arg.EventType+netType, mgr)
	}
	updateExcelParam(mgr, arg)
	err = mgr.InsertExcelSql()
	if err != nil {
		log.Errorf("InsertExcelSql error: %s", err)
		return writeResponse(ctx, ResponsePack(InsertSqlError))
	}
	res := ResponsePack(SUCCESS)
	res["Result"] = arg
	return writeResponse(ctx, res)
}

func GetAdminBalanceByEventType(ctx *routing.Context) error {
	evtType := ctx.Param("evtty")
	boo := ledger.DefBonusLedger.HasExcelEvtTy(evtType)
	if !boo {
		return writeResponse(ctx, ResponsePack(NotExistenceEvtType))
	}
	netType := ctx.Param("netty")
	mgr, errCode := parseMgr(evtType, netType)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	adminBalance, err := mgr.GetAdminBalance()
	if err != nil {
		log.Errorf("GetAdminBalance error: %s", err)
		return writeResponse(ctx, ResponsePack(GetAdminBalanceError))
	}
	res := ResponsePack(SUCCESS)
	res["Result"] = adminBalance
	return writeResponse(ctx, res)
}

func GetGasPrice(ctx *routing.Context) error {
	netty := ctx.Param("netty")
	evtty := ctx.Param("evtty")
	mgr, errCode := parseMgr(evtty, netty)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	res := ResponsePack(SUCCESS)
	res["Result"] = mgr.GetGasPrice()
	return writeResponse(ctx, res)
}

func SetGasPrice(ctx *routing.Context) error {
	gasPriceInt, evtTy, netTy, errCode := ParseSetGasPriceParam(ctx)
	if errCode != SUCCESS {
		log.Errorf("ParseSetGasPriceParam error ")
		return writeResponse(ctx, ResponsePack(errCode))
	}
	mgr, err := parseMgr(evtTy, netTy)
	if err != SUCCESS {
		return writeResponse(ctx, ResponsePack(err))
	}
	mgr.SetGasPrice(int(gasPriceInt))
	ledger.DefBonusLedger.SetGasPrice(evtTy, netTy, int(gasPriceInt))
	return writeResponse(ctx, ResponsePack(SUCCESS))
}

func Transfer(ctx *routing.Context) error {
	eventType, netType, errCode := ParseTransferParam(ctx)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	boo := ledger.DefBonusLedger.HasExcelEvtTy(eventType)
	if !boo {
		return writeResponse(ctx, ResponsePack(NotExistenceEvtType))
	}
	mgr, errCode := parseMgr(eventType, netType)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	log.Info("transfer status:", mgr.GetStatus())
	if mgr.GetStatus() == common.Transfering {
		return writeResponse(ctx, ResponsePack(Transfering))
	}
	if !ledger.DefBonusLedger.HasTxInfoEvtTy(eventType) {
		ledger.DefBonusLedger.AppendTxInfoEvtTy(eventType)
	}
	mgr.StartTransfer()
	log.Info("start transfer success")
	return writeResponse(ctx, ResponsePack(SUCCESS))
}

func Withdraw(ctx *routing.Context) error {
	withdrawParam, errCode := ParseWithdrawParam(ctx)
	if errCode != SUCCESS || withdrawParam.EventType == "" || withdrawParam.NetType == "" ||
		withdrawParam.TokenType == "" || withdrawParam.Address == "" {
			log.Errorf("[Withdraw] ParseWithdrawParam error:%s", errCode)
		return writeResponse(ctx, ResponsePack(errCode))
	}
	boo := ledger.DefBonusLedger.HasExcelEvtTy(withdrawParam.EventType)
	if !boo {
		return writeResponse(ctx, ResponsePack(NotExistenceEvtType))
	}
	mgr, errCode := parseMgr(withdrawParam.EventType, withdrawParam.NetType)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}

	if mgr.VerifyAddress(withdrawParam.Address) == false {
		return writeResponse(ctx, ResponsePack(AddressIsWrong))
	}
	log.Info("transfer status:", mgr.GetStatus())
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

func GetExcelEventType(ctx *routing.Context) error {
	res := ResponsePack(SUCCESS)
	res["Result"] = ledger.DefBonusLedger.AllEvtTys.AllExcelEvtTy
	return writeResponse(ctx, res)
}
func GetTxInfoEventType(ctx *routing.Context) error {
	netTy := ctx.Param("netty")
	if netTy != config.TestNet && netTy != config.MainNet {
		return writeResponse(ctx, ResponsePack(NetTypeError))
	}
	res := ResponsePack(SUCCESS)
	res["Result"] = ledger.DefBonusLedger.AllEvtTys.AllTxInfoEvtTy
	return writeResponse(ctx, res)
}

func GetTransferProgress(ctx *routing.Context) error {
	evtty := ctx.Param("evtty")
	boo := ledger.DefBonusLedger.HasTxInfoEvtTy(evtty)
	if !boo {
		return writeResponse(ctx, ResponsePack(NotExistenceEvtType))
	}
	netty := ctx.Param("netty")
	mgr, errCode := parseMgr(evtty, netty)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	res, err := mgr.GetDB().QueryTransferProgress(evtty, netty)
	if err != nil {
		log.Errorf("[GetTransferProgress] QueryTransferProgress failed: %s", err)
		return writeResponse(ctx, ResponsePack(QueryTransferProgressFailed))
	}
	total := mgr.GetTotal()
	res["total"] = total
	r := ResponsePack(SUCCESS)
	r["Result"] = res
	return writeResponse(ctx, r)
}

func GetExcelParamByEvtType(ctx *routing.Context) error {
	param, errCode := ParseQueryExcelParam(ctx)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	boo := ledger.DefBonusLedger.HasExcelEvtTy(param.EvtType)
	if !boo {
		return writeResponse(ctx, ResponsePack(NotExistenceEvtType))
	}
	mgr, errCode := parseMgr(param.EvtType, param.NetType)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	var excelParam *common.ExcelParam
	var err error
	if param.PageSize == 0 && param.PageNum == 0 {
		excelParam, err = mgr.GetDB().QueryExcelParamByEventType(param.EvtType, 0, 0)
	} else {
		if param.PageNum <= 0 {
			param.PageNum = 1
		}
		start := (param.PageNum - 1) * param.PageSize
		end := start + param.PageSize
		fmt.Println(mgr.GetDB())
		excelParam, err = mgr.GetDB().QueryExcelParamByEventType(param.EvtType, start, end)
	}

	if err != nil {
		log.Errorf("QueryExcelParamByEventType error:%s", err)
		return writeResponse(ctx, ResponsePack(errCode))
	}

	updateExcelParam(mgr, excelParam)
	res := ResponsePack(SUCCESS)
	res["Result"] = excelParam
	return writeResponse(ctx, res)
}

func updateExcelParam(mgr interfaces.WithdrawManager, excelParam *common.ExcelParam) int64 {
	var err error
	excelParam.Sum, err = mgr.ComputeSum()
	if err != nil {
		log.Errorf("InitManager error: %s", err)
		return SumError
	}
	excelParam.AdminBalance, err = mgr.GetAdminBalance()
	if err != nil {
		log.Errorf("GetAdminBalance error: %s", err)
		return GetAdminBalanceError
	}
	excelParam.EstimateFee, err = mgr.EstimateFee("", mgr.GetTotal())
	if err != nil {
		log.Errorf("EstimateFee error: %s", err)
		return EstimateFeeError
	}
	excelParam.Admin = mgr.GetAdminAddress()
	excelParam.Total = mgr.GetTotal()
	return SUCCESS
}

func GetTxInfoByEventType(ctx *routing.Context) error {
	param, errCode := ParseQueryTxInfoParam(ctx)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	boo := ledger.DefBonusLedger.HasTxInfoEvtTy(param.EvtTy)
	if !boo {
		return writeResponse(ctx, ResponsePack(NotExistenceEvtType))
	}
	mgr, errCode := parseMgr(param.EvtTy, param.NetTy)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	if param.PageNum < 1 {
		param.PageNum = 1
	}
	start := (param.PageNum - 1) * param.PageSize
	end := start + param.PageSize
	txInfo, err := mgr.GetDB().QueryTxInfoByEventType(param.EvtTy, start, end)
	if err != nil {
		log.Errorf("QueryTxInfoByEventType error: %s", err)
		return writeResponse(ctx, ResponsePack(QueryResultByEventType))
	}

	sum, err := mgr.ComputeSum()
	if err != nil {
		log.Errorf("InitManager error: %s", err)
		return writeResponse(ctx, ResponsePack(SumError))
	}
	ba, err := mgr.GetAdminBalance()
	if err != nil {
		log.Errorf("GetAdminBalance error: %s", err)
		return writeResponse(ctx, ResponsePack(GetAdminBalanceError))
	}
	fee, err := mgr.EstimateFee("", mgr.GetTotal())
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
	}

	r := ResponsePack(SUCCESS)
	r["Result"] = res
	return writeResponse(ctx, r)
}

func parseMgr(eventType, netType string) (interfaces.WithdrawManager, int64) {
	var mgr interfaces.WithdrawManager
	mn, ok := DefBonusMap.Load(eventType + netType)

	//TODO
	if !ok || mn == nil {
		var err error
		mgr, err = manager.RecoverManager(eventType, netType)
		if err != nil {
			log.Errorf("InitManager error: %s", err)
			return nil, InitManagerError
		}
		DefBonusMap.Store(eventType+netType, mgr)
	} else {
		mgr, ok = mn.(interfaces.WithdrawManager)
		if !ok {
			return nil, TypeTransferError
		}
	}
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
