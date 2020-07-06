package manager

import (
	"fmt"

	"github.com/ontio/bonus/bonus_db"
	"github.com/ontio/bonus/common"
	"github.com/ontio/bonus/config"
	"github.com/ontio/bonus/manager/eth"
	"github.com/ontio/bonus/manager/interfaces"
	"github.com/ontio/bonus/manager/ont"
)

func CreateManager(eatp *common.ExcelParam, netType string) (interfaces.WithdrawManager, error) {
	if eatp.TokenType == config.OEP4 || eatp.TokenType == config.ERC20 {
		if eatp.ContractAddress == "" {
			return nil, fmt.Errorf("ContractAddress is nil")
		}
	}
	var mgr interfaces.WithdrawManager
	var err error
	switch eatp.TokenType {
	case config.ONG, config.OEP4, config.ONT:
		// init ont mgr in working-path
		mgr, err = ont.NewOntManager(config.DefConfig.OntCfg, eatp, netType)
		if err != nil {
			return nil, err
		}
	case config.ERC20, config.ETH:
		// init eth mgr in working-path
		mgr, err = eth.NewEthManager(config.DefConfig.EthCfg, eatp, netType)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("[createManager] no support token, tokenType: %s", eatp.TokenType)
	}
	for _, tr := range eatp.BillList {
		if !mgr.VerifyAddress(tr.Address) {
			return nil, fmt.Errorf("address is wrong: %s", tr.Address)
		}
	}
	return mgr, nil
}

func RecoverManager(tokenType, eventType, netType string) (interfaces.WithdrawManager, error) {
	db, err := bonus_db.NewBonusDB(tokenType, eventType, netType)
	if err != nil {
		return nil, err
	}
	switch tokenType {
	case config.ONG, config.OEP4, config.ONT:
		// init ont mgr in working-path
		_, eatp, err := ont.LoadOntManager(tokenType, eventType)
		if err != nil {
			return nil, fmt.Errorf("failed to load ont mgr: %s", err)
		}
		ontManager, err := ont.NewOntManager(config.DefConfig.OntCfg, eatp, netType)
		if err != nil {
			return nil, err
		}
		ontManager.SetDB(db)
		return ontManager, nil
	case config.ERC20, config.ETH:
		// init eth mgr in working-path
		_, eatp, err := eth.LoadEthManager(tokenType, eventType)
		if err != nil {
			return nil, fmt.Errorf("failed to load eth mgr: %s", err)
		}
		ethManager, err := eth.NewEthManager(config.DefConfig.EthCfg, eatp, netType)
		if err != nil {
			return nil, err
		}
		ethManager.SetDB(db)
		return ethManager, nil
	}
	return nil, fmt.Errorf("[createManager] no support token, tokenType: %s, eventType: %s", tokenType, eventType)
}
