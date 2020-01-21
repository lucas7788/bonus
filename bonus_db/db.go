package bonus_db

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/ontio/bonus/common"
	"github.com/ontio/bonus/config"
	"strings"
)

var DefDB *sql.DB

func ConnectDB() error {
	db, dberr := sql.Open("mysql",
		config.DefConfig.BonusDBUser+
			":"+config.DefConfig.BonusDBPassword+
			"@tcp("+config.DefConfig.BonusDBUrl+
			")/"+config.DefConfig.BonusDBName+
			"?charset=utf8")
	if dberr != nil {
		return dberr
	}
	err := db.Ping()
	if err != nil {
		return err
	}
	DefDB = db
	return nil
}

func CloseDB() {
	DefDB.Close()
}

func QueryAllEventType() ([]string, error) {
	strSql := "select EventType from bonus_transaction_info"
	stmt, err := DefDB.Prepare(strSql)
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

func InsertSql(args *common.ExcelParam) error {
	sqlStrArr := make([]string, 0)
	for _, bill := range args.BillList {
		txInfo, err := QueryTxHexByExcelAndAddr(args.EventType, bill.Address)
		if err != nil {
			return err
		}
		if txInfo != nil {
			continue
		}
		oneData := fmt.Sprintf("('%s','%s','%s','%s','%s')", args.EventType, args.TokenType, args.ContractAddress, bill.Address, bill.Amount)
		sqlStrArr = append(sqlStrArr, oneData)
	}
	if len(sqlStrArr) == 0 {
		return fmt.Errorf("database has the same data")
	}
	content := strings.Join(sqlStrArr, ",")
	strSql := "insert into bonus_transaction_info (EventType,TokenType,ContractAddress, Address, Amount) values"
	fmt.Println(strSql + content)
	_, err := DefDB.Exec(strSql + content)
	if err != nil {
		return err
	}
	return nil
}

func UpdateTxInfo(txHash, TxHex string, txResult common.TxResult, eventType, address string) error {
	strSql := "update bonus_transaction_info set TxHash=?,TxHex=?,TxResult=? where EventType = ? and Address = ?"
	stmt, err := DefDB.Prepare(strSql)
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		return err
	}
	_, err = stmt.Exec(txHash, TxHex, txResult, eventType, address)
	return err
}

func UpdateTxResult(eventType, address string, txResult common.TxResult, txTime uint32, errDetail string) error {
	strSql := "update bonus_transaction_info set TxResult=?, TxTime=?, ErrorDetail= ? where EventType = ? and Address = ?"
	stmt, err := DefDB.Prepare(strSql)
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		return err
	}
	_, err = stmt.Exec(txResult, txTime, errDetail, eventType, address)
	return err
}

func QueryTxHexByExcelAndAddr(eventType, address string) (*common.TransactionInfo, error) {
	strSql := "select TxHash,TxHex,TxResult from bonus_transaction_info where EventType=? and Address=?"
	stmt, err := DefDB.Prepare(strSql)
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(eventType, address)
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

func QueryResultByEventType(eventType string, res []*common.TransactionInfo) ([]*common.TransactionInfo, error) {
	strSql := "select TokenType,ContractAddress,Address,Amount,TxHash,TxTime,TxResult from bonus_transaction_info where EventType = ?"
	stmt, err := DefDB.Prepare(strSql)
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
	for rows.Next() {
		var txHash, tokenType, contractAddress, address, amount string
		var txResult byte
		var txTime uint32
		if err = rows.Scan(&tokenType, &contractAddress, &address, &amount, &txHash, &txTime, &txResult); err != nil {
			return nil, err
		}
		res = append(res, &common.TransactionInfo{
			EventType:       eventType,
			TokenType:       tokenType,
			ContractAddress: contractAddress,
			Address:         address,
			Amount:          amount,
			TxHash:          txHash,
			TxTime:          txTime,
			TxResult:        common.TxResult(txResult),
		})
	}
	return res, nil
}
