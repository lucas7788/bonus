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

func CreateManager(eatp *common.ExcelParam, netType string, db *bonus_db.BonusDB) (interfaces.WithdrawManager, error) {
	if db == nil {
		var err error
		db, err = bonus_db.NewBonusDB(eatp.TokenType, eatp.EventType, netType)
		if err != nil {
			return nil, err
		}
	}
	if eatp.TokenType == config.OEP4 || eatp.TokenType == config.ERC20 {
		if eatp.ContractAddress == "" {
			return nil, fmt.Errorf("ContractAddress is nil")
		}
	}
	switch eatp.TokenType {
	case config.ONG, config.OEP4, config.ONT:
		// init ont mgr in working-path
		ontManager, err := ont.NewOntManager(config.DefConfig.OntCfg, eatp, netType, db)
		if err != nil {
			return nil, err
		}
		return ontManager, nil
	case config.ERC20, config.ETH:
		// init eth mgr in working-path
		ethManager, err := eth.NewEthManager(config.DefConfig.EthCfg, eatp, netType, db)
		if err != nil {
			return nil, err
		}
		return ethManager, nil
	default:
		return nil, fmt.Errorf("[createManager] no support token, tokenType: %s", eatp.TokenType)
	}
}

func RecoverManager(tokenType, eventType, netType string) (interfaces.WithdrawManager, error) {
	db, err := bonus_db.NewBonusDB(tokenType, eventType, netType)
	if err != nil {
		return nil, err
	}
	switch tokenType {
	case config.ONG, config.OEP4, config.ONT:
		// init ont mgr in working-path
		cfg, eatp, err := ont.LoadOntManager(tokenType, eventType)
		if err != nil {
			return nil, fmt.Errorf("failed to load ont mgr: %s", err)
		}
		ontManager, err := ont.NewOntManager(cfg, eatp, netType, db)
		if err != nil {
			return nil, err
		}
		return ontManager, nil
	case config.ERC20, config.ETH:
		// init eth mgr in working-path
		cfg, eatp, err := eth.LoadEthManager(tokenType, eventType)
		if err != nil {
			return nil, fmt.Errorf("failed to load eth mgr: %s", err)
		}
		ethManager, err := eth.NewEthManager(cfg, eatp, netType, db)
		if err != nil {
			return nil, err
		}
		return ethManager, nil
	}
	return nil, fmt.Errorf("[createManager] no support token, tokenType: %s, eventType: %s", tokenType, eventType)
}
