package bonus_db

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
)

type AllEventType struct {
	EvtTypes []string
	db       *leveldb.DB
}

func (this *AllEventType) IsContains(evtType string) bool {
	for _, item := range this.EvtTypes {
		if item == evtType {
			return true
		}
	}
	return false
}

func (this *AllEventType) AddEvtType(evtType string) error {
	this.EvtTypes = append(this.EvtTypes, evtType)
	sink := common.NewZeroCopySink(nil)
	this.Serialize(sink)
	return this.db.Put([]byte(config.EventType), sink.Bytes(), nil)
}

func InitAllEventType() (*AllEventType, error) {
	db, err := InitLevelDB(config.EventType)
	if err != nil {
		return nil, err
	}
	val, err := db.Get([]byte(config.EventType), nil)
	if err != nil {
		return nil, err
	}
	aet := &AllEventType{}
	source := common.NewZeroCopySource(val)
	err = aet.Deserialize(source)
	if err != nil {
		return nil, err
	}
	return aet, nil
}

func (this *AllEventType) Serialize(sink *common.ZeroCopySink) {
	sink.WriteUint32(uint32(len(this.EvtTypes)))
	for _, item := range this.EvtTypes {
		sink.WriteString(item)
	}
}

func (this *AllEventType) Deserialize(source *common.ZeroCopySource) error {
	l, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	res := make([]string, l)
	for i := 0; i < int(l); i++ {
		data, err := readStr(source)
		if err != nil {
			return err
		}
		res[i] = data
	}
	return nil
}

func readStr(source *common.ZeroCopySource) (string, error) {
	data, _, irregular, eof := source.NextString()
	if irregular {
		return "", common.ErrIrregularData
	}
	if eof {
		return "", io.ErrUnexpectedEOF
	}
	return data, nil
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
func InitLevelDB(eventType string) (*leveldb.DB, error) {
	path := fmt.Sprintf("%s%s%s", config.DefConfig.LevelDBPath, string(os.PathSeparator), "db_"+eventType)
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
