package restful

import (
	"encoding/json"
	"github.com/ontio/bonus/common"
	"github.com/ontio/bonus/config"
	"github.com/ontio/ontology/common/log"
	"github.com/qiangxue/fasthttp-routing"
)


func ParseExcelAndTransferParam(ctx *routing.Context) (*common.ExcelAndTransferParam, int64) {
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
	tokenType, ok := para["tokenType"].(string)
	if !ok || tokenType == "" {
		log.Errorf("tokenType error\n")
		return nil, PARA_ERROR
	}
	contractAddress, ok := para["contractAddress"].(string)
	if !ok {
		log.Errorf("contractAddress error\n")
		return nil, PARA_ERROR
	}
	privateKey, ok := para["privateKey"].(string)
	if !ok {
		log.Errorf("privateKey error\n")
		return nil, PARA_ERROR
	}
	if tokenType == config.ERC20 && (privateKey == "" || contractAddress == "") {
		log.Errorf("privateKey error\n")
		return nil, PARA_ERROR
	}
	walletFile, ok := para["walletFile"].(string)
	if !ok {
		log.Errorf("walletFile error\n")
		return nil, PARA_ERROR
	}
	pwd, ok := para["pwd"].(string)
	if !ok {
		log.Errorf("pwd error\n")
		return nil, PARA_ERROR
	}
	if tokenType == config.OEP4 && (walletFile == "" || pwd == "" || contractAddress == "") {
		log.Errorf("param error\n")
		return nil, PARA_ERROR
	}

	fileName, ok := para["fileName"].(string)
	if !ok || fileName == "" {
		log.Errorf("fileName error\n")
		return nil, PARA_ERROR
	}
	transferParam := make([]*common.TransferParam, 0)
	billListRaw, ok := para["billList"]
	if !ok {
		log.Errorf("param Unmarshal error: %s \n", err)
		return nil, PARA_ERROR
	}
	billList, ok := billListRaw.([]interface{})
	if !ok {
		log.Errorf("billList error\n")
		return nil, PARA_ERROR
	}
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
		amt, ok := pi["amount"].(float64)
		if !ok {
			log.Errorf("amount parse error,")
			log.Info("address", pi["amount"])
			return nil, PARA_ERROR
		}
		tp := &common.TransferParam{
			Address: addr,
			Amount:  float64(amt),
		}
		transferParam = append(transferParam, tp)
	}
	return &common.ExcelAndTransferParam{
		BillList:        transferParam,
		TokenType:       tokenType,
		PrivateKey:      privateKey,
		ContractAddress: contractAddress,
		FileName:        fileName,
		WalletFileContent:      walletFile,
		Pwd:             pwd,
	}, SUCCESS
}
