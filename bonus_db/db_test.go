package bonus_db

import (
	"testing"

	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/ontio/bonus/common"
	"github.com/stretchr/testify/assert"
)

func TestInsertSql(t *testing.T) {
	db, err := NewBonusDB()
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	tp := &common.TransferParam{
		Id:      0,
		Address: "111",
		Amount:  "111",
	}
	Num := 10000
	tps := make([]*common.TransferParam, Num)
	tps = append(tps, tp)
	for i := 0; i < Num; i++ {
		tps[i] = tp
	}
	ep := &common.ExcelParam{
		BillList:        tps,
		TokenType:       "11",
		ContractAddress: "111",
		EventType:       "111",
		Admin:           "111",
		EstimateFee:     "111",
		Sum:             "string",
		AdminBalance:    nil,
	}

	db.InsertExcelSql(ep)
	evts, err := db.QueryAllEventType()
	assert.Nil(t, err)
	fmt.Println(evts)
}

func TestBonusDB_InsertTxInfoSql(t *testing.T) {
	db, err := NewBonusDB()
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
	db.InsertTxInfoSql([]*common.TransactionInfo{txInfo})
	info, err := db.QueryTxHexByExcelAndAddr("11", "11", 1)
	assert.Nil(t, err)
	fmt.Println(info)
}
