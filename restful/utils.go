package restful

import (
	"encoding/json"
	"github.com/ontio/bonus/common"
	"github.com/ontio/bonus/config"
	"github.com/ontio/ontology/common/log"
	"github.com/qiangxue/fasthttp-routing"
	"strings"
)

func ParseQueryDataParam(ctx *routing.Context) ([]string, int64) {
	param, errCode := parse(ctx)
	if errCode != SUCCESS {
		return nil, errCode
	}
	arg, ok := param["eventType"]
	if !ok {
		return nil, PARA_ERROR
	}
	pa, ok := arg.([]interface{})
	if !ok {
		return nil, PARA_ERROR
	}
	res := make([]string, 0)
	for _, item := range pa {
		it, ok := item.(string)
		if !ok {
			return nil, TypeTransferError
		}
		res = append(res, it)
	}
	return res, SUCCESS
}

func ParseTransferParam(ctx *routing.Context) (string, int64) {
	param, errCode := parse(ctx)
	if errCode != SUCCESS {
		return "", errCode
	}
	args, ok := param["eventType"]
	if !ok {
		return "", PARA_ERROR
	}
	pa, ok := args.(string)
	if !ok {
		return "", PARA_ERROR
	}
	return pa, SUCCESS
}
func ParseWithdrawParam(ctx *routing.Context) (string, string, int64) {
	param, errCode := parse(ctx)
	if errCode != SUCCESS {
		return "", "", errCode
	}
	eventType, ok := param["eventType"]
	if !ok {
		return "", "", PARA_ERROR
	}
	evtType, ok := eventType.(string)
	if !ok {
		return "", "", PARA_ERROR
	}
	address, ok := param["address"]
	if !ok {
		return "", "", PARA_ERROR
	}
	addr, ok := address.(string)
	if !ok {
		return "", "", PARA_ERROR
	}
	return evtType, addr, SUCCESS
}

func parse(ctx *routing.Context) (map[string]interface{}, int64) {
	req := ctx.PostBody()
	if req == nil || len(req) == 0 {
		log.Errorf("param length is 0\n")
		return nil, PARA_ERROR
	}
	arg := make(map[string]interface{})
	err := json.Unmarshal(req, &arg)
	if err != nil {
		log.Errorf("param Unmarshal error: %s \n", err)
		return nil, PARA_PARSE_ERROR
	}
	param, ok := arg["params"]
	para, ok := param.(map[string]interface{})
	if !ok {
		log.Errorf("param.(map[string]interface{})\n")
		return nil, PARA_ERROR
	}
	return para, SUCCESS
}

func ParseExcelParam(ctx *routing.Context) (*common.ExcelParam, int64) {
	param, errCode := parse(ctx)
	if errCode != SUCCESS {
		return nil, errCode
	}

	tokenType, ok := param["tokenType"].(string)
	if !ok || tokenType == "" {
		log.Errorf("tokenType error\n")
		return nil, PARA_ERROR
	}
	contractAddress, ok := param["contractAddress"].(string)
	if !ok {
		log.Errorf("contractAddress error\n")
		return nil, PARA_ERROR
	}
	if tokenType == config.ERC20 && contractAddress == "" {
		log.Errorf("tokenType == config.ERC20 and contractAddress is nil\n")
		return nil, PARA_ERROR
	}
	if tokenType == config.OEP4 && contractAddress == "" {
		log.Errorf("param error\n")
		return nil, PARA_ERROR
	}
	eventType, ok := param["eventType"].(string)
	if !ok || eventType == "" {
		log.Errorf("transfer is nil\n")
		return nil, PARA_ERROR
	}
	transferParam := make([]*common.TransferParam, 0)
	billListRaw, ok := param["billList"]
	if !ok {
		log.Errorf("param Unmarshal error\n")
		return nil, PARA_ERROR
	}
	billList, ok := billListRaw.([]interface{})
	if !ok {
		log.Errorf("billList error\n")
		return nil, PARA_ERROR
	}
	tempAddr := make([]string, 0)
	for _, p := range billList {
		pi, ok := p.(map[string]interface{})
		if !ok {
			log.Errorf("p error\n")
			return nil, PARA_ERROR
		}
		addr, ok := pi["address"].(string)
		if !ok {
			log.Errorf("address parse error,")
			log.Info("address", pi["address"])
			return nil, PARA_ERROR
		}
		if common.IsHave(tempAddr, addr) {
			return nil, ExcelDuplicateAddress
		}
		amt, ok := pi["amount"].(string)
		if !ok {
			log.Errorf("amount parse error,")
			log.Info("address", pi["amount"])
			return nil, AmountIsNegative
		}
		if strings.Contains(amt, "-") {
			log.Errorf("amount have -,")
			log.Info("address", pi["amount"])
			return nil, PARA_ERROR
		}
		tp := &common.TransferParam{
			Address: addr,
			Amount:  amt,
		}
		transferParam = append(transferParam, tp)
	}
	return &common.ExcelParam{
		BillList:        transferParam,
		TokenType:       tokenType,
		ContractAddress: contractAddress,
		EventType:       eventType,
	}, SUCCESS
}

func ParseTxInfoToEatp(txInfo []*common.TransactionInfo) *common.ExcelParam {
	billList := make([]*common.TransferParam, 0)
	for _, item := range txInfo {
		billList = append(billList, &common.TransferParam{
			Id:      item.Id,
			Address: item.Address,
			Amount:  item.Amount,
		})
	}
	res := &common.ExcelParam{
		BillList:        billList,
		TokenType:       txInfo[0].TokenType,
		EventType:       txInfo[0].EventType,
		ContractAddress: txInfo[0].ContractAddress,
	}
	return res
}
