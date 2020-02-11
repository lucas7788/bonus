package restful

import (
	"encoding/json"
	"github.com/ontio/bonus/bonus_db"
	"github.com/ontio/bonus/common"
	"github.com/ontio/bonus/manager"
	"github.com/ontio/bonus/manager/interfaces"
	"github.com/ontio/ontology/common/log"
	"github.com/qiangxue/fasthttp-routing"
	"strconv"
	"sync"
)

var DefBonusMap *sync.Map //projectId -> Airdrop

func UpLoadExcel(ctx *routing.Context) error {
	arg, errCode := ParseExcelParam(ctx)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
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
	var err error
	if !hasInit {
		mgr, err = manager.InitManager(arg)
		if err != nil {
			log.Errorf("InitManager error: %s", err)
			return writeResponse(ctx, ResponsePack(InitManagerError))
		}
		DefBonusMap.Store(arg.EventType, mgr)
	}
	arg.Sum, err = mgr.ComputeSum()
	if err != nil {
		log.Errorf("ComputeSum error: %s", err)
		res := ResponsePack(SumError)
		res["Result"] = err.Error()
		return writeResponse(ctx, res)
	}
	arg.AdminBalance, err = mgr.GetAdminBalance()
	if err != nil {
		log.Errorf("GetAdminBalance error: %s", err)
		return writeResponse(ctx, ResponsePack(InitManagerError))
	}
	arg.EstimateFee, err = mgr.EstimateFee()
	if err != nil {
		log.Errorf("EstimateFee error, err: %s", err)
		return writeResponse(ctx, ResponsePack(EstimateFeeError))
	}
	err = bonus_db.InsertSql(arg)
	if err != nil {
		log.Errorf("InsertSql error: %s", err)
		return writeResponse(ctx, ResponsePack(InsertSqlError))
	}

	arg.Admin = mgr.GetAdminAddress()
	res := ResponsePack(SUCCESS)
	res["Result"] = arg
	return writeResponse(ctx, res)
}

func Transfer(ctx *routing.Context) error {
	eventType, errCode := ParseTransferParam(ctx)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	mgr, errCode := parseMgr(eventType)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	log.Info("transfer status:", mgr.GetStatus())
	if mgr.GetStatus() == common.Transfering {
		return writeResponse(ctx, ResponsePack(Transfering))
	}
	sum, _ := mgr.ComputeSum()
	sumF, _ := strconv.ParseFloat(sum, 64)
	balance, _ := mgr.GetAdminBalance()
	balanceF, _ := strconv.ParseFloat(balance, 64)
	if balanceF < sumF {
		return writeResponse(ctx, ResponsePack(BalanceIsNotEnough))
	}
	mgr.StartTransfer()
	log.Info("start transfer success")
	return writeResponse(ctx, ResponsePack(SUCCESS))
}

func Withdraw(ctx *routing.Context) error {
	eventType, address, errCode := ParseWithdrawParam(ctx)
	if errCode != SUCCESS || eventType == "" || address == "" {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	mgr, errCode := parseMgr(eventType)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	if mgr.VerifyAddress(address) == false {
		return writeResponse(ctx, ResponsePack(AddressIsWrong))
	}
	log.Info("transfer status:", mgr.GetStatus())
	if mgr.GetStatus() == common.Transfering {
		return writeResponse(ctx, ResponsePack(Transfering))
	}
	err := mgr.WithdrawToken(address)
	if err != nil {
		log.Errorf("WithdrawToken failed, error: %s", err)
		res := ResponsePack(WithdrawTokenFailed)
		res["Result"] = err
		return writeResponse(ctx, res)
	}
	return writeResponse(ctx, ResponsePack(SUCCESS))
}

func GetAllEventType(ctx *routing.Context) error {
	eventType, err := bonus_db.QueryAllEventType()
	if err != nil {
		log.Errorf("QueryAllEventTypeError error: %s", err)
		return writeResponse(ctx, ResponsePack(QueryAllEventTypeError))
	}
	res := ResponsePack(SUCCESS)
	res["Result"] = eventType
	return writeResponse(ctx, res)
}

func GetAdminBalanceByEventType(ctx *routing.Context) error {
	eventType := ctx.Param("eventtype")
	mgr, errCode := parseMgr(eventType)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	balance, err := mgr.GetAdminBalance()
	if err != nil {
		log.Errorf("GetAdminBalance error: %s", err)
		return writeResponse(ctx, ResponsePack(GetAdminBalanceError))
	}
	res := ResponsePack(SUCCESS)
	res["Result"] = balance
	return writeResponse(ctx, res)
}

func parseMgr(eventType string) (interfaces.WithdrawManager, int64) {
	var mgr interfaces.WithdrawManager
	mn, ok := DefBonusMap.Load(eventType)
	//TODO
	if !ok || mn == nil || true {
		res := make([]*common.TransactionInfo, 0)
		var err error
		res, err = bonus_db.QueryResultByEventType(eventType, res)
		if err != nil {
			log.Errorf("there is no the eventType: %s", eventType)
			return nil, NoTheEventTypeError
		}
		mgr, err = manager.InitManager(ParseTxInfoToEatp(res))
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

func GetDataByEventType(ctx *routing.Context) error {
	eventType, errCode := ParseQueryDataParam(ctx)
	if errCode != SUCCESS {
		return writeResponse(ctx, ResponsePack(errCode))
	}
	txInfo := make([]*common.TransactionInfo, 0)
	res := make([]*common.GetDataByEventType, 0)
	for _, ty := range eventType {
		var err error
		txInfo, err = bonus_db.QueryResultByEventType(ty, txInfo)
		if err != nil {
			log.Errorf("QueryResultByEventType error: %s", err)
			return writeResponse(ctx, ResponsePack(QueryResultByEventType))
		}
		eatp := ParseTxInfoToEatp(txInfo)
		mgr, err := manager.InitManager(eatp)
		if err != nil {
			log.Errorf("InitManager error: %s", err)
			return writeResponse(ctx, ResponsePack(InitManagerError))
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
		fee, err := mgr.EstimateFee()
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
