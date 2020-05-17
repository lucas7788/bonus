package bonus_db

import (
	"testing"

	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/ontio/bonus/common"
	"github.com/ontio/bonus/config"
	"github.com/stretchr/testify/assert"
)

func init() {

}

func TestBonusDB_InsertTxInfoSql(t *testing.T) {
	db, err := NewBonusDB(config.ONG, "eeee5", config.TestNet)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	txInfo := &common.TransactionInfo{
		Id:        1,
		NetType:   "testNet",
		EventType: "11",
		Address:   "11",
	}
	txInfo2 := &common.TransactionInfo{
		Id:        2,
		NetType:   "testNet",
		EventType: "11",
		Address:   "11",
	}
	err = db.InsertTxInfoSql([]*common.TransactionInfo{txInfo, txInfo2})
	assert.Nil(t, err)
	info, err := db.QueryTxHexByExcelAndAddr("11", "11", 1)
	assert.Nil(t, err)
	fmt.Println(info)

	sum, err := db.QueryTxInfoNum()
	assert.Nil(t, err)
	fmt.Println("sum:", sum)
}
