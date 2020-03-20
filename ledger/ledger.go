package ledger

import (
	"fmt"
	"github.com/ontio/bonus/config"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"io"
	"os"
	"strconv"
)

//excel_evt_ty -> allExcelEvtType
//tx_info_evt_ty -> all txInfo EvtType
//evtty_netty -> gasprice
var DefBonusLedger *BonusLedger

type BonusLedger struct {
	db        *leveldb.DB
	AllEvtTys *AllEvtTy
}

func (this *BonusLedger) GetGasPrice(evtTy, netTy string) (int, error) {
	gasBs, err := this.db.Get([]byte(config.KEY_GAS_PRICE+evtTy+netTy), nil)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(gasBs))
}

func (this *BonusLedger) SetGasPrice(evtTy, netTy string, gasPrice int) error {
	err := this.db.Put([]byte(config.KEY_GAS_PRICE+evtTy+netTy), []byte(strconv.Itoa(gasPrice)), nil)
	if err != nil {
		log.Errorf("SetGasPrice error:%s", err)
		return err
	}
	return nil
}

func (this *BonusLedger) Close() {
	this.db.Close()
}

func NewBonusLedger() (*BonusLedger, error) {
	db, err := InitLevelDB()
	if err != nil {
		return nil, err
	}

	bl := &BonusLedger{
		db: db,
		AllEvtTys: &AllEvtTy{
			AllExcelEvtTy:  make([]string, 0),
			AllTxInfoEvtTy: make([]string, 0),
		},
	}
	err = bl.init()
	if err != nil {
		return nil, err
	}
	return bl, nil
}

type AllEvtTy struct {
	AllExcelEvtTy  []string
	AllTxInfoEvtTy []string
}

func (this *AllEvtTy) Serialize(sink *common.ZeroCopySink) {
	sink.WriteUint32(uint32(len(this.AllExcelEvtTy)))
	for i := 0; i < len(this.AllExcelEvtTy); i++ {
		sink.WriteString(this.AllExcelEvtTy[i])
	}
	sink.WriteUint32(uint32(len(this.AllTxInfoEvtTy)))
	for i := 0; i < len(this.AllTxInfoEvtTy); i++ {
		sink.WriteString(this.AllTxInfoEvtTy[i])
	}
}

func (this *AllEvtTy) Deserialize(source *common.ZeroCopySource) error {
	excel, err := read(source)
	if err != nil {
		return err
	}
	txInfo, err := read(source)
	if err != nil {
		return err
	}
	this.AllExcelEvtTy = excel
	this.AllTxInfoEvtTy = txInfo
	return nil
}

func read(source *common.ZeroCopySource) ([]string, error) {
	l, eof := source.NextUint32()
	if eof {
		return nil, io.ErrUnexpectedEOF
	}
	strs := make([]string, l)
	for i := 0; i < int(l); i++ {
		var ir, eof bool
		strs[i], _, ir, eof = source.NextString()
		if ir {
			return nil, common.ErrIrregularData
		}
		if eof {
			return nil, io.ErrUnexpectedEOF
		}
	}
	return strs, nil
}

func (this *BonusLedger) init() error {
	val, err := this.db.Get([]byte(config.KEY_EVT_TY), nil)
	if err == leveldb.ErrNotFound {
		return nil
	} else if err != nil {
		return err
	}
	source := common.NewZeroCopySource(val)
	evts := &AllEvtTy{
		AllExcelEvtTy:  make([]string, 0),
		AllTxInfoEvtTy: make([]string, 0),
	}
	err = evts.Deserialize(source)
	if err != nil {
		return err
	}
	this.AllEvtTys = evts
	return nil
}

func (this *BonusLedger) HasExcelEvtTy(evtTy string) bool {
	for _, ty := range this.AllEvtTys.AllExcelEvtTy {
		if ty == evtTy {
			return true
		}
	}
	return false
}

func (this *BonusLedger) AppendExcelEvtTy(evtTy string) {
	this.AllEvtTys.AllExcelEvtTy = append(this.AllEvtTys.AllExcelEvtTy, evtTy)
	this.updateDB()
}

func (this *BonusLedger) updateDB() {
	sink := common.NewZeroCopySink(nil)
	this.AllEvtTys.Serialize(sink)
	err := this.db.Put([]byte(config.KEY_EVT_TY), sink.Bytes(), nil)
	if err != nil {
		log.Errorf("[AppendExcelEvtTy] error: %s", err)
		return
	}
}

func (this *BonusLedger) AppendTxInfoEvtTy(evtTy string) {
	this.AllEvtTys.AllTxInfoEvtTy = append(this.AllEvtTys.AllTxInfoEvtTy, evtTy)
	this.updateDB()
}

func (this *BonusLedger) HasTxInfoEvtTy(evtTy string) bool {
	for _, ty := range this.AllEvtTys.AllTxInfoEvtTy {
		if ty == evtTy {
			return true
		}
	}
	return false
}

func OpenLevelDB(file string) (*leveldb.DB, error) {
	openFileCache := opt.DefaultOpenFilesCacheCapacity

	// default Options
	o := opt.Options{
		NoSync:                 false,
		OpenFilesCacheCapacity: openFileCache,
		Filter:                 filter.NewBloomFilter(10),
	}

	db, err := leveldb.OpenFile(file, &o)

	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		db, err = leveldb.RecoverFile(file, nil)
	}
	if err != nil {
		return nil, err
	}
	return db, nil
}

func InitLevelDB() (*leveldb.DB, error) {
	path := fmt.Sprintf("%s%s%s", config.DefConfig.LevelDBPath, string(os.PathSeparator), "leveldb")
	db, err := OpenLevelDB(path)
	if err != nil {
		return nil, err
	}
	log.Infof("ledger init success")
	return db, nil
}
func CloseLevelDB(db *leveldb.DB) {
	if db != nil {
		db.Close()
	}
}
