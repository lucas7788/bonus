package core

import (
	"github.com/ontio/bonus/bonus_db"
	"github.com/ontio/bonus/common"
	"github.com/ontio/bonus/manager"
	"github.com/ontio/bonus/manager/interfaces"
	"github.com/ontio/ontology/common/log"
	"github.com/syndtr/goleveldb/leveldb"
)

type Bonus struct {
	db        *leveldb.DB
	txManager interfaces.WithdrawManager
}

func NewBonus(arg *common.ExcelParam) (*Bonus, error) {
	mgr, err := manager.InitManager(arg)
	if err != nil {
		log.Errorf("InitManager error: %s", err)
		return nil, err
	}
	db, err := bonus_db.InitLevelDB(arg.EventType)
	if err != nil {
		return nil, err
	}
	bonus := &Bonus{
		db:        db,
		txManager: mgr,
	}
	err = bonus.UpdateArgs(arg)
	if err != nil {
		return nil, err
	}
	return bonus, nil
}

func (this *Bonus) UpdateDataFromDB() {

}

func (this *Bonus) UpdateArgs(arg *common.ExcelParam) error {
	var err error
	arg.Sum, err = this.txManager.ComputeSum()
	if err != nil {
		log.Errorf("ComputeSum error: %s", err)
		return err
	}
	arg.AdminBalance, err = this.txManager.GetAdminBalance()
	if err != nil {
		log.Errorf("GetAdminBalance error: %s", err)
		return err
	}
	arg.EstimateFee, err = this.txManager.EstimateFee()
	if err != nil {
		log.Errorf("EstimateFee error, err: %s", err)
		return err
	}

	arg.Admin = this.txManager.GetAdminAddress()
	return nil
}