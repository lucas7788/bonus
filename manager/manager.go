package manager

import (
	"fmt"
	"github.com/ontio/bonus/bonus_db"
	"github.com/ontio/bonus/common"
	"github.com/ontio/bonus/config"
	"github.com/ontio/bonus/ledger"
	"github.com/ontio/bonus/manager/eth"
	"github.com/ontio/bonus/manager/interfaces"
	"github.com/ontio/bonus/manager/ont"
	"github.com/ontio/ontology/common/log"
)

func InitManager(eatp *common.ExcelParam, netType string, db *bonus_db.BonusDB) (interfaces.WithdrawManager, error) {
	if db == nil {
		var err error
		db, err = bonus_db.NewBonusDB(eatp.EventType, netType)
		if err != nil {
			return nil, err
		}
	}
	if eatp.TokenType == config.OEP4 || eatp.TokenType == config.ERC20 {
		if eatp.ContractAddress == "" {
			return nil, fmt.Errorf("ContractAddress is nil")
		}
	manager, err := createManager(eatp, netType, db)
	if err != nil {
		return nil, err
	}
	//set contract address
	if eatp.TokenType == config.OEP4 || eatp.TokenType == config.ERC20 {
		err = manager.SetContractAddress(eatp.ContractAddress)
		if err != nil {
			log.Errorf("manager SetContractAddress failed, error: %s", err)
			return nil, fmt.Errorf("SetContractAddress failed, error: %s", err)
		}
	}
	return manager, nil
}

func RecoverManager(evtTy, netTy string) (interfaces.WithdrawManager, error) {
	db, err := bonus_db.NewBonusDB(evtTy, netTy)
	if err != nil {
		return nil, err
	}
	excelParam, err := db.QueryExcelParamByEventType(evtTy, 0, 0)
	if err != nil {
		db.Close()
		return nil, err
	}
	mgr, err := InitManager(excelParam, netTy, db)
	if err != nil {
		return nil, err
	}
	gasPrice, _ := ledger.DefBonusLedger.GetGasPrice(evtTy, netTy)
	if gasPrice != 0 {
		mgr.SetGasPrice(gasPrice)
	}
	return mgr, err
}

func createManager(eatp *common.ExcelParam, netType string, db *bonus_db.BonusDB) (interfaces.WithdrawManager, error) {

	switch eatp.TokenType {
	case config.ONG, config.OEP4, config.ONT:
		//init ont manager
		ontManager, err := ont.NewOntManager(config.DefConfig.OntCfg, eatp, netType, db)
		if err != nil {
			return nil, err
		}
		return ontManager, nil
	case config.ERC20, config.ETH:
		ethManager, err := eth.NewEthManager(config.DefConfig.EthCfg, eatp, netType, db)
		if err != nil {
			return nil, err
		}
		return ethManager, nil
	default:
		return nil, fmt.Errorf("[createManager] no support token, tokenType: %s", eatp.TokenType)
	}
}
