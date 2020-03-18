package restful

import (
	"encoding/json"
	"github.com/ontio/bonus/bonus_db"
	"github.com/ontio/bonus/common"
	"github.com/ontio/bonus/manager"
	"github.com/ontio/bonus/manager/eth"
	"github.com/ontio/bonus/manager/interfaces"
	"github.com/ontio/ontology/common/log"
	"github.com/qiangxue/fasthttp-routing"
	"math/big"
	"strings"
	"sync"
)

var DefBonusMap *sync.Map //projectId -> Airdrop

func UpLoadExcel(ctx *routing.Context) error {
	arg, errCode := ParseExcelParam(ctx)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	evtTypes, err := bonus_db.DefBonusDB.QueryAllEventType()
	if err != nil {
		log.Errorf("QueryAllEventType error: %s", err)
		return writeResponse(ctx, ResponsePack(QueryAllEventTypeError))
	}
	if common.IsHave(evtTypes, arg.EventType) {
		log.Errorf("DuplicateEventType: %s", arg.EventType)
		return writeResponse(ctx, ResponsePack(DuplicateEventType))
	}
	hasInit := false
	var mgr interfaces.WithdrawManager
	if DefBonusMap == nil {
		DefBonusMap = new(sync.Map)
	} else {
		mn, ok := DefBonusMap.Load(arg.EventType)
		if ok && mn != nil {
			mgr, _ = mn.(interfaces.WithdrawManager)
			hasInit = true
		}
	}
	if !hasInit {
		mgr, err = manager.InitManager(arg, arg.NetType)
		if err != nil {
			log.Errorf("InitManager error: %s", err)
			return writeResponse(ctx, ResponsePack(InitManagerError))
		}
		DefBonusMap.Store(arg.EventType+arg.NetType, mgr)
	}
	updateExcelParam(mgr, arg)
	err = bonus_db.DefBonusDB.InsertExcelSql(arg)
	if err != nil {
		log.Errorf("InsertExcelSql error: %s", err)
		return writeResponse(ctx, ResponsePack(InsertSqlError))
	}
	res := ResponsePack(SUCCESS)
	res["Result"] = arg
	return writeResponse(ctx, res)
}

func GetAdminBalanceByEventType(ctx *routing.Context) error {
	evtType := ctx.Param("eventtype")
	netType := ctx.Param("nettype")
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
	gasPrice := new(big.Int).Div(eth.DEFAULT_GAS_PRICE, eth.OneGwei)
	res := ResponsePack(SUCCESS)
	res["Result"] = gasPrice.Uint64()
	return writeResponse(ctx, res)
}

func SetGasPrice(ctx *routing.Context) error {
	gasPriceInt, errCode := ParseSetGasPriceParam(ctx)
	if errCode != SUCCESS {
		log.Errorf("ParseSetGasPriceParam error ")
		return writeResponse(ctx, ResponsePack(errCode))
	}
	gasPrice := new(big.Int).SetUint64(uint64(gasPriceInt))
	eth.DEFAULT_GAS_PRICE = new(big.Int).Mul(gasPrice, eth.OneGwei)
	return writeResponse(ctx, ResponsePack(SUCCESS))
}

func Transfer(ctx *routing.Context) error {
	eventType, netType, errCode := ParseTransferParam(ctx)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	mgr, errCode := parseMgr(eventType, netType)
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
	if errCode != SUCCESS || withdrawParam.EventType == "" || withdrawParam.NetType == "" ||
		withdrawParam.TokenType == "" || withdrawParam.Address == "" {
		return writeResponse(ctx, ResponsePack(errCode))
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

func GetAllEventType(ctx *routing.Context) error {
	eventType, err := bonus_db.DefBonusDB.QueryAllEventType()
	if err != nil {
		log.Errorf("QueryAllEventTypeError error: %s", err)
		return writeResponse(ctx, ResponsePack(QueryAllEventTypeError))
	}
	res := ResponsePack(SUCCESS)
	res["Result"] = eventType
	return writeResponse(ctx, res)
}

func GetTransferProgress(ctx *routing.Context) error {
	evtType, netType, errCode := ParseTransferParam(ctx)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	res, err := bonus_db.DefBonusDB.QueryTransferProgress(evtType, netType)
	if err != nil {
		return writeResponse(ctx, ResponsePack(QueryTransferProgressFailed))
	}
	mgr, errCode := parseMgr(evtType, netType)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	total := mgr.GetTotal()
	res["total"] = total
	r := ResponsePack(SUCCESS)
	r["Result"] = res
	return writeResponse(ctx, r)
}

func GetExcelParamByEvtType(ctx *routing.Context) error {
	evtType, netType, errCode := ParseTransferParam(ctx)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	excelParam, err := bonus_db.DefBonusDB.QueryExcelParamByEventType(evtType, netType)
	if err != nil {
		log.Errorf("QueryExcelParamByEventType error:%s", err)
		return writeResponse(ctx, ResponsePack(errCode))
	}
	mgr, errCode := parseMgr(evtType, netType)
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
	return SUCCESS
}

func GetTxInfoByEventType(ctx *routing.Context) error {
	netTy, eventType, errCode := ParseQueryDataParam(ctx)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	txInfo := make([]*common.TransactionInfo, 0)
	res := make([]*common.GetDataByEventType, 0)

	for _, ty := range eventType {
		var err error
		txInfo, err = bonus_db.DefBonusDB.QueryTxInfoByEventType(ty)
		if err != nil {
			log.Errorf("QueryTxInfoByEventType error: %s", err)
			return writeResponse(ctx, ResponsePack(QueryResultByEventType))
		}
		eatp := ParseTxInfoToEatp(txInfo)
		mn, ok := DefBonusMap.Load(eatp.EventType)
		var mgr interfaces.WithdrawManager
		if ok && mn != nil {
			mgr, _ = mn.(interfaces.WithdrawManager)
		} else {
			mgr, err = manager.InitManager(eatp, netTy)
			if err != nil {
				log.Errorf("InitManager error: %s", err)
				return writeResponse(ctx, ResponsePack(InitManagerError))
			}
			DefBonusMap.Store(eatp.EventType, mgr)
		}
		eatp.Sum, err = mgr.ComputeSum()
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
		res = append(res, &common.GetDataByEventType{
			TxInfo:          txInfo,
			AdminBalance:    ba,
			Admin:           mgr.GetAdminAddress(),
			Sum:             eatp.Sum,
			EstimateFee:     fee,
			EventType:       txInfo[0].EventType,
			TokenType:       txInfo[0].TokenType,
			ContractAddress: txInfo[0].ContractAddress,
		})
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
		res, err := bonus_db.DefBonusDB.QueryExcelParamByEventType(eventType, netType)
		if err != nil {
			log.Errorf("there is no the eventType: %s, err: %s", eventType, err)
			return nil, NoTheEventTypeError
		}
		mgr, err = manager.InitManager(res, netType)
		if err != nil {
			log.Errorf("InitManager error: %s", err)
			return nil, InitManagerError
		}

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
