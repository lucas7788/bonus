package manager

import (
	"fmt"
	"github.com/ontio/bonus/config"
	"github.com/ontio/bonus/manager/eth"
	"github.com/ontio/bonus/manager/interfaces"
	"github.com/ontio/bonus/manager/ont"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/bonus/common"
)

func InitManager(eatp *common.ExcelAndTransferParam) (interfaces.WithdrawManager, error) {
	manager, err := createManager(eatp)
	if err != nil {
		return nil, err
	}
	//set contract address
	if eatp.TokenType == config.OEP4 || eatp.TokenType == config.ERC20 {
		if eatp.ContractAddress == "" {
			log.Error("ContractAddress is nil")
			return nil, fmt.Errorf("ContractAddress is nil")
		}
		err = manager.SetContractAddress(eatp.ContractAddress)
		if err != nil {
			log.Errorf("manager SetContractAddress failed, error: %s", err)
			return nil, fmt.Errorf("SetContractAddress failed, error: %s", err)
		}
	}
	return manager, nil
}

func createManager(eatp *common.ExcelAndTransferParam) (interfaces.WithdrawManager, error) {

	switch eatp.TokenType {
	case config.ONG, config.OEP4, config.ONT:
		//init ont manager
		ontManager, err := ont.NewOntManager(config.DefConfig.Ont, eatp)
		if err != nil {
			return nil, err
		}
		return ontManager, nil
	case config.ERC20:

		ethManager, err := eth.NewEthManager(config.DefConfig.EthCfg, eatp)
		if err != nil {
			return nil, err
		}
		return ethManager, nil
	default:
		return nil, fmt.Errorf("no support token, tokenType: %s", eatp.TokenType)
	}
}
