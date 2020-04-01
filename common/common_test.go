package common

import (
	"fmt"
	"testing"
)

func TestClearData(t *testing.T) {
	netty := "netty"
	tokenty := "tokenty"
	evtTy := "evtTy"
	dbFileName := GetEventDBFilename(netty, tokenty, evtTy)
	if err := CheckPath(dbFileName); err != nil {
		return
	}
	ClearData(netty, tokenty, evtTy)
}

func TestExcelParam_TrParamSort(t *testing.T) {
	billList := make([]*TransferParam, 0)
	for i := 10; i > 0; i-- {
		tp := &TransferParam{
			Id: i,
		}
		billList = append(billList, tp)
	}

	ep := &ExcelParam{
		BillList: billList,
	}
	fmt.Println(ep.BillList[0].Id)
	ep.TrParamSort()
	fmt.Println(ep.BillList[0].Id)
}
