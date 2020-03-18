package bonus_db

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/ontio/bonus/common"
	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"os"
	"strings"
)

var DefBonusDB *BonusDB

type BonusDB struct {
	db *sql.DB
}

func NewBonusDB() (*BonusDB, error) {
	dbFileName := "./db/" + "bonus" + ".db"
	if !common2.FileExisted("./db") {
		err := os.Mkdir("./db", os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	if !common2.FileExisted(dbFileName) {
		file, err := os.Create(dbFileName)
		if err != nil {
			return nil, err
		}
		file.Close()
	}

	db, dberr := sql.Open("sqlite3", dbFileName)
	if dberr != nil {
		return nil, dberr
	}
	err := db.Ping()
	if err != nil {
		return nil, err
	}
	createTxInfoTableSql := `CREATE TABLE IF NOT EXISTS "bonus_transaction_info"("Id" INTEGER PRIMARY KEY NOT NULL, "NetType" varchar(20) not null, "TokenType" varchar(20) not null, "EventType" varchar(100) not null, "ContractAddress" varchar(100) not null, "Address" varchar(100) not null, "Amount" varchar(100) not null, "TxHash" varchar(100) not null DEFAULT "", "TxTime" bigint(20) NOT NULL DEFAULT 0, "TxHex" varchar(5000) not null default "", "ErrorDetail" varchar(1000) not null default "", "TxResult" tinyint(1) NOT NULL DEFAULT 0)`
	_, err = db.Exec(createTxInfoTableSql, nil)
	if err != nil {
		return nil, err
	}
	createExcelTableSql := `CREATE TABLE IF NOT EXISTS "excel_info"("Id" INTEGER PRIMARY KEY autoincrement NOT NULL, "TokenType" varchar(20) not null, "EventType" varchar(100) not null,"NetType" varchar(15) not null, "ContractAddress" varchar(100) not null, "Address" varchar(100) not null, "Amount" varchar(100) not null)`
	_, err = db.Exec(createExcelTableSql, nil)
	if err != nil {
		return nil, err
	}
	return &BonusDB{
		db: db,
	}, nil
}

func (this *BonusDB) Close() {
	this.db.Close()
}

func (this *BonusDB) QueryAllEventType() ([]string, error) {
	strSql := "select EventType from excel_info"
	stmt, err := this.db.Prepare(strSql)
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query()
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return nil, err
	}
	res := make([]string, 0)
	for rows.Next() {
		var eventType string
		if err = rows.Scan(&eventType); err != nil {
			return nil, err
		}
		if !common.IsHave(res, eventType) {
			res = append(res, eventType)
		}
	}
	return res, nil
}

func (this *BonusDB) InsertExcelSql(args *common.ExcelParam) error {
	sqlStrArr := make([]string, 0)
	for _, bill := range args.BillList {
		oneData := fmt.Sprintf("('%s','%s','%s','%s','%s', '%s')", args.NetType, args.EventType, args.TokenType, args.ContractAddress, bill.Address, bill.Amount)
		sqlStrArr = append(sqlStrArr, oneData)
	}
	if len(sqlStrArr) == 0 {
		return fmt.Errorf("database has the same data")
	}
	content := strings.Join(sqlStrArr, ",")
	strSql := "insert into excel_info (NetType,EventType,TokenType,ContractAddress, Address, Amount) values"
	_, err := this.db.Exec(strSql + content)
	if err != nil {
		return err
	}
	err = this.UpdateId(args)
	if err != nil {
		log.Errorf("UpdateId error: %s, eventType: %s", err, args.EventType)
		return err
	}
	return nil
}

func (this *BonusDB) InsertTxInfoSql(args []*common.TransactionInfo) error {
	sqlStrArr := make([]string, 0)
	for _, txInfo := range args {
		oneData := fmt.Sprintf("('%d','%s','%s','%s','%s','%s','%s','%s','%s','%d','%s')", txInfo.Id, txInfo.NetType, txInfo.EventType, txInfo.TokenType, txInfo.ContractAddress, txInfo.Address, txInfo.Amount, txInfo.TxHash, txInfo.TxHex, txInfo.TxResult, txInfo.ErrorDetail)
		sqlStrArr = append(sqlStrArr, oneData)
	}
	if len(sqlStrArr) == 0 {
		return fmt.Errorf("database has the same data")
	}
	content := strings.Join(sqlStrArr, ",")
	strSql := "insert into bonus_transaction_info (Id,NetType,EventType,TokenType,ContractAddress, Address, Amount,TxHash,TxHex,TxResult,ErrorDetail) values"
	_, err := this.db.Exec(strSql + content)
	if err != nil {
		return err
	}
	return nil
}

func (this *BonusDB) UpdateId(args *common.ExcelParam) error {
	strSql := "select Id,Address,Amount from excel_info where EventType=?"
	stmt, err := this.db.Prepare(strSql)
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		return err
	}
	rows, err := stmt.Query(args.EventType)
	if rows != nil {
		defer rows.Close()
	}
	tps := make([]*common.TransferParam, 0)
	for rows.Next() {
		var id int
		var addr, amount string
		if err = rows.Scan(&id, &addr, &amount); err != nil {
			return err
		}
		tps = append(tps, &common.TransferParam{
			Id:      id,
			Address: addr,
			Amount:  amount,
		})
	}
	args.BillList = tps
	return nil
}

func (this *BonusDB) UpdateTxInfo(txHash, TxHex string, txResult common.TxResult, eventType, address string, id int) error {
	strSql := "update bonus_transaction_info set TxHash=?,TxHex=?,TxResult=? where EventType = ? and Address = ? and id = ?"
	stmt, err := this.db.Prepare(strSql)
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		return err
	}
	_, err = stmt.Exec(txHash, TxHex, txResult, eventType, address, id)
	return err
}

func (this *BonusDB) UpdateTxResult(eventType, address string, id int, txResult common.TxResult, txTime uint32, errDetail string) error {
	strSql := "update bonus_transaction_info set TxResult=?, TxTime=?, ErrorDetail= ? where EventType = ? and Address = ? and id=?"
	stmt, err := this.db.Prepare(strSql)
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		return err
	}
	_, err = stmt.Exec(txResult, txTime, errDetail, eventType, address, id)
	return err
}

func (this *BonusDB) QueryTransferProgress(eventType, netType string) (map[string]int, error) {
	success, err := this.getSum(eventType, netType, common.TxSuccess)
	if err != nil {
		return nil, err
	}
	failed, err := this.getSum(eventType, netType, common.TxFailed)
	if err != nil {
		return nil, err
	}
	transfering, err := this.getSum(eventType, netType, common.OneTransfering)
	if err != nil {
		return nil, err
	}
	notSend, err := this.getSum(eventType, netType, common.NotSend)
	if err != nil {
		return nil, err
	}
	sendFailed, err := this.getSum(eventType, netType, common.SendFailed)
	if err != nil {
		return nil, err
	}
	res := make(map[string]int)
	res["success"] = success
	res["failed"] = failed
	res["transfering"] = transfering
	res["notSend"] = notSend
	res["sendFailed"] = sendFailed
	return res, nil
}

func (this *BonusDB) getSum(eventType, netType string, txResult common.TxResult) (int, error) {
	strSql := "select sum(Id) from bonus_transaction_info where EventType=? and netType=? and TxResult=?"
	stmt, err := this.db.Prepare(strSql)
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		return 0, err
	}
	rows, err := stmt.Query(eventType, netType, common.TxSuccess)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return 0, err
	}
	for rows.Next() {
		var sum int
		if err = rows.Scan(&sum); err != nil {
			return 0, err
		}
		return sum, nil
	}
	return 0, nil
}

func (this *BonusDB) QueryTxHexByExcelAndAddr(eventType, address string, id int) (*common.TransactionInfo, error) {
	strSql := "select TxHash,TxHex,TxResult from bonus_transaction_info where EventType=? and Address=? and Id=?"
	stmt, err := this.db.Prepare(strSql)
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(eventType, address, id)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var txHash, txHex string
		var txResult int
		if err = rows.Scan(&txHash, &txHex, &txResult); err != nil {
			return nil, err
		}
		return &common.TransactionInfo{
			TxHash:   txHash,
			TxHex:    txHex,
			TxResult: common.TxResult(txResult),
		}, nil
	}
	return nil, nil
}

func (this *BonusDB) QueryExcelParamByEventType(evtType, netType string) (*common.ExcelParam, error) {
	strSql := "select Id, TokenType,ContractAddress,Address,Amount from excel_info where EventType = ? and NetType=?"
	stmt, err := this.db.Prepare(strSql)
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(evtType, netType)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return nil, err
	}
	tokenTy := ""
	contractAddr := ""
	billList := make([]*common.TransferParam, 0)
	for rows.Next() {
		var tokenType, contractAddress, address, amount string
		var id int
		if err = rows.Scan(&id, &tokenType, &contractAddress, &address, &amount); err != nil {
			return nil, err
		}
		if contractAddress != "" && contractAddr == "" {
			tokenTy = tokenType
			contractAddr = contractAddress
		}
		billList = append(billList, &common.TransferParam{
			Id:      id,
			Address: address,
			Amount:  amount,
		})
	}
	return &common.ExcelParam{
		BillList:        billList,
		TokenType:       tokenTy,
		ContractAddress: contractAddr,
		EventType:       evtType,
	}, nil
}

func (this *BonusDB) QueryTxInfoByEventType(eventType string) ([]*common.TransactionInfo, error) {
	strSql := "select Id, TokenType,ContractAddress,Address,Amount,TxHash,TxTime,TxResult,ErrorDetail from bonus_transaction_info where EventType = ?"
	stmt, err := this.db.Prepare(strSql)
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(eventType)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return nil, err
	}
	res := make([]*common.TransactionInfo, 0)
	for rows.Next() {
		var txHash, tokenType, contractAddress, address, amount, errorDetail string
		var txResult byte
		var txTime uint32
		var id int
		if err = rows.Scan(&id, &tokenType, &contractAddress, &address, &amount, &txHash, &txTime, &txResult, &errorDetail); err != nil {
			return nil, err
		}
		res = append(res, &common.TransactionInfo{
			Id:              id,
			EventType:       eventType,
			TokenType:       tokenType,
			ContractAddress: contractAddress,
			Address:         address,
			Amount:          amount,
			TxHash:          txHash,
			TxTime:          txTime,
			TxResult:        common.TxResult(txResult),
			ErrorDetail:     errorDetail,
		})
	}
	return res, nil
}
