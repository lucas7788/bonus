package bonus_db

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ontio/bonus/common"
	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
)

//var DefBonusDB *BonusDB

type BonusDB struct {
	db *sql.DB
}

func NewBonusDB(tokenTy, evtTy, netTy string) (*BonusDB, error) {
	dbFileName := common.GetEventDBFilename(netTy, tokenTy, evtTy)
	if err := common.CheckPath(dbFileName); err != nil {
		return nil, fmt.Errorf("failed to create %s: %s", dbFileName, err)
	}
	existed := true
	if !common2.FileExisted(dbFileName) {
		existed = false
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
	log.Info("[NewBonusDB] existed:", existed)
	if !existed {
		createTxInfoTableSqlTest := `CREATE TABLE IF NOT EXISTS "bonus_transaction_info"("Id" INTEGER PRIMARY KEY NOT NULL, "NetType" varchar(20) not null, "TokenType" varchar(20) not null, "EventType" varchar(100) not null, "ContractAddress" varchar(100) not null, "Address" varchar(100) not null, "Amount" varchar(100) not null, "TxHash" varchar(100) not null DEFAULT "", "TxTime" bigint(20) NOT NULL DEFAULT 0, "TxHex" varchar(5000) not null default "", "ErrorDetail" varchar(1000) not null default "", "TxResult" tinyint(1) NOT NULL DEFAULT 0)`
		_, err = db.Exec(createTxInfoTableSqlTest, nil)
		if err != nil {
			return nil, err
		}
	}
	return &BonusDB{
		db: db,
	}, nil
}

func (this *BonusDB) Close() {
	this.db.Close()
}

func (this *BonusDB) InsertTxInfoSql(args []*common.TransactionInfo) error {
	sqlStrArr := make([]string, 0)
	for _, txInfo := range args {
		if txInfo == nil {
			continue
		}
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

func (this *BonusDB) UpdateTxResultByTxHash(txHash string, txResult common.TxResult, txTime uint32, errDetail string) error {
	strSql := "update bonus_transaction_info set TxResult=?, TxTime=?, ErrorDetail= ? where TxHash=?"
	stmt, err := this.db.Prepare(strSql)
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		return err
	}
	_, err = stmt.Exec(txResult, txTime, errDetail, txHash)
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
	notSend, err := this.getSum(eventType, netType, common.NotBuild)
	if err != nil {
		return nil, err
	}
	sendFailed, err := this.getSum(eventType, netType, common.SendFailed)
	if err != nil {
		return nil, err
	}
	total, err := this.getSum(eventType, netType, common.AllStatus)
	if err != nil {
		return nil, err
	}
	res := make(map[string]int)
	res["success"] = success
	res["failed"] = failed
	res["transfering"] = transfering
	res["notSend"] = notSend
	res["sendFailed"] = sendFailed
	res["total"] = total
	return res, nil
}

func (this *BonusDB) getSum(eventType, netType string, txResult common.TxResult) (int, error) {
	var strSql string
	if txResult == common.AllStatus {
		strSql = "select count(Id) from bonus_transaction_info where EventType=? and NetType=?"
	} else {
		strSql = "select count(Id) from bonus_transaction_info where EventType=? and NetType=? and TxResult=?"
	}

	stmt, err := this.db.Prepare(strSql)
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		return 0, err
	}
	var rows *sql.Rows
	if txResult == common.AllStatus {
		rows, err = stmt.Query(eventType, netType)
	} else {
		rows, err = stmt.Query(eventType, netType, txResult)
	}
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

func (this *BonusDB) QueryTxHexByTxHash(txHash string) (*common.TransactionInfo, error) {
	strSql := "select TxHash,TxHex,TxResult from bonus_transaction_info where TxHash=?"

	stmt, err := this.db.Prepare(strSql)
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(txHash)
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

func (this *BonusDB) QueryTxInfo(txCaches map[string]*common.TxCache) (map[string]*common.TxCache, error) {
	strSql := "select Id,Address,TxHash,TxHex,TxResult from bonus_transaction_info"
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
	for rows.Next() {
		var txHash, txHex, addr string
		var txResult common.TxResult
		if err = rows.Scan(&addr, &txHash, &txHex, &txResult); err != nil {
			return nil, err
		}
		txHexBytes, err := common2.HexToBytes(txHex)
		if err != nil {
			panic(err)
		}
		txCaches[addr] = &common.TxCache{
			Addr:     addr,
			TxHash:   txHash,
			TxHex:    txHexBytes,
			TxStatus: txResult,
		}
	}
	return txCaches, nil
}

func (this *BonusDB) QueryTxResult(addr string) (common.TxResult, error) {
	strSql := "select TxResult from bonus_transaction_info where Address=?"
	stmt, err := this.db.Prepare(strSql)
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		return common.NotBuild, err
	}
	rows, err := stmt.Query(addr)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return common.NotBuild, err
	}
	for rows.Next() {
		var txResult common.TxResult
		if err = rows.Scan(&txResult); err != nil {
			return common.NotBuild, err
		}
		return txResult, nil
	}
	return common.NotBuild, nil
}

func (this *BonusDB) QueryTxInfoNum() (int, error) {
	strSql := "select count(*) from bonus_transaction_info"
	stmt, err := this.db.Prepare(strSql)
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		return 0, err
	}
	rows, err := stmt.Query()
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return 0, err
	}
	for rows.Next() {
		var txInfoNum int
		if err = rows.Scan(&txInfoNum); err != nil {
			return 0, err
		}
		return txInfoNum, nil
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

func (this *BonusDB) QueryTxInfoByEventType(eventType string, start, pageSize int, txResult common.TxResult) ([]*common.TransactionInfo, int, error) {
	var strSql string
	if txResult == common.AllStatus {
		if start == 0 && pageSize == 0 {
			strSql = "select Id, TokenType,NetType,ContractAddress,Address,Amount,TxHash,TxTime,TxResult,ErrorDetail from bonus_transaction_info where EventType = ?"
		} else {
			strSql = "select Id, TokenType,NetType,ContractAddress,Address,Amount,TxHash,TxTime,TxResult,ErrorDetail from bonus_transaction_info where EventType = ?  order by id limit ?, ?"
		}
	} else {
		if start == 0 && pageSize == 0 {
			strSql = "select Id, TokenType,NetType,ContractAddress,Address,Amount,TxHash,TxTime,TxResult,ErrorDetail from bonus_transaction_info where EventType = ? and TxResult = ?"
		} else {
			strSql = "select Id, TokenType,NetType,ContractAddress,Address,Amount,TxHash,TxTime,TxResult,ErrorDetail from bonus_transaction_info where EventType = ? and TxResult = ? order by id limit ?, ?"
		}
	}
	stmt, err := this.db.Prepare(strSql)
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		return nil, 0, err
	}
	var rows *sql.Rows
	if txResult == common.AllStatus {
		if start == 0 && pageSize == 0 {
			rows, err = stmt.Query(eventType)
		} else {
			rows, err = stmt.Query(eventType, start, pageSize)
		}
	} else {
		if start == 0 && pageSize == 0 {
			rows, err = stmt.Query(eventType, txResult)
		} else {
			rows, err = stmt.Query(eventType, txResult, start, pageSize)
		}
	}
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return nil, 0, err
	}
	res := make([]*common.TransactionInfo, 0)
	for rows.Next() {
		var txHash, tokenType, netty, contractAddress, address, amount, errorDetail string
		var txResult byte
		var txTime uint32
		var id int
		if err = rows.Scan(&id, &tokenType, &netty, &contractAddress, &address, &amount, &txHash, &txTime, &txResult, &errorDetail); err != nil {
			return nil, 0, err
		}
		res = append(res, &common.TransactionInfo{
			Id:              id,
			EventType:       eventType,
			TokenType:       tokenType,
			NetType:         netty,
			ContractAddress: contractAddress,
			Address:         address,
			Amount:          amount,
			TxHash:          txHash,
			TxTime:          txTime,
			TxResult:        common.TxResult(txResult),
			ErrorDetail:     errorDetail,
		})
	}
	total, err := this.getTxInfoTotal(eventType, txResult)
	if err != nil {
		return nil, 0, err
	}
	return res, total, nil
}

func (this *BonusDB) getTxInfoTotal(eventType string, txResult common.TxResult) (int, error) {
	var strSql string
	if txResult == common.AllStatus {
		strSql = "select count(Id) from bonus_transaction_info where EventType=?"
	} else {
		strSql = "select count(Id) from bonus_transaction_info where EventType=? and TxResult=?"
	}

	stmt, err := this.db.Prepare(strSql)
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		return 0, err
	}
	var rows *sql.Rows
	if txResult == common.AllStatus {
		rows, err = stmt.Query(eventType)
		if err != nil {
			return 0, err
		}
	} else {
		rows, err = stmt.Query(eventType, txResult)
		if err != nil {
			return 0, err
		}
	}

	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return 0, err
	}
	for rows.Next() {
		var total int
		if err = rows.Scan(&total); err != nil {
			return 0, err
		}
		return total, nil
	}
	return 0, nil
}
