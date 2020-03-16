package common

import (
	"io"
	"github.com/ontio/ontology/common"
)

type TransferParam struct {
	Id      int
	Address string
	Amount  string
}

type TransactionInfo struct {
	Id              int
	EventType       string
	TokenType       string
	ContractAddress string
	Address         string
	Amount          string
	TxHash          string
	TxTime          uint32
	TxHex           string
	TxResult        TxResult
	ErrorDetail     string
}

type GetDataByEventType struct {
	TxInfo          []*TransactionInfo
	Admin           string
	EstimateFee     string
	Sum             string
	AdminBalance    string
	EventType       string
	TokenType       string
	ContractAddress string
}

type TransferStatus byte

const (
	NotTransfer TransferStatus = iota
	Transfering
	Transfered
)

type TxResult byte

const (
	NotBuild TxResult = iota //
	BuildTxFailed
	SendFailed
	OneTransfering
	TxFailed
	TxSuccess
)

type ExcelParam struct {
	BillList        []*TransferParam
	TokenType       string
	ContractAddress string
	EventType       string
	Admin           string
	EstimateFee     string
	Sum             string
	AdminBalance    map[string]string
}

type BonusInfo struct {
	EventType       string
	TokenType       string
	ContractAddress string
	TxInfos         []*TxInfo
	Admin           string
	EstimateFee     string
	Sum             string
	AdminBalance    map[string]string
}

func (this *BonusInfo) Serialize(sink *common.ZeroCopySink) {
	sink.WriteString(this.EventType)
	sink.WriteString(this.TokenType)
	sink.WriteString(this.ContractAddress)
	sink.WriteUint32(uint32(len(this.TxInfos)))
	for _, txInfo := range this.TxInfos {
		txInfo.Serialize(sink)
	}
}

func (this *BonusInfo) Deserialize(source *common.ZeroCopySource) error {
	var err error
	this.EventType, err = readStr(source)
	if err != nil {
		return err
	}
	this.TokenType, err = readStr(source)
	if err != nil {
		return err
	}
	this.ContractAddress, err = readStr(source)
	if err != nil {
		return err
	}
	l, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	txInfos := make([]*TxInfo, l)
	for i := 0; i < int(l); i++ {
		txInfo := &TxInfo{}
		err = txInfo.Deserialize(source)
		if err != nil {
			return err
		}
		txInfos[i] = txInfo
	}
	this.TxInfos = txInfos
	return nil
}

func ParseExcelParamToBonusInfo(param *ExcelParam) *BonusInfo {
	txInfos := make([]*TxInfo, len(param.BillList))
	for i := 0; i < len(param.BillList); i++ {
		txInfo := &TxInfo{
			Id:      param.BillList[i].Id,
			Amount:  param.BillList[i].Amount,
			Address: param.BillList[i].Address,
		}
		txInfos[i] = txInfo
	}
	return &BonusInfo{
		EventType:       param.EventType,
		TokenType:       param.TokenType,
		ContractAddress: param.ContractAddress,
		TxInfos:         txInfos,
		Admin:param.Admin,
		EstimateFee:param.EstimateFee,
		Sum:param.Sum,
		AdminBalance:    param.AdminBalance,
	}
}

type TxInfo struct {
	Id          int
	Address     string
	Amount      string
	TxHash      string
	TxHex       string
	TxResult    TxResult
	ErrorDetail string
	TxTime      uint32
}

func (this *TxInfo) Serialize(sink *common.ZeroCopySink) {
	sink.WriteUint32(uint32(this.Id))
	sink.WriteString(this.Address)
	sink.WriteString(this.Amount)
	sink.WriteString(this.TxHash)
	sink.WriteString(this.TxHex)
	sink.WriteByte(byte(this.TxResult))
	sink.WriteString(this.ErrorDetail)
	sink.WriteUint32(this.TxTime)
}

func (this *TxInfo) Deserialize(source *common.ZeroCopySource) error {
	id, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Id = int(id)
	var err error
	this.Address, err = readStr(source)
	if err != nil {
		return err
	}
	this.Amount, err = readStr(source)
	if err != nil {
		return err
	}
	this.TxHash, err = readStr(source)
	if err != nil {
		return err
	}
	this.TxHex, err = readStr(source)
	if err != nil {
		return err
	}
	txRes, eof := source.NextByte()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.TxResult = TxResult(txRes)
	this.ErrorDetail, err = readStr(source)
	if err != nil {
		return err
	}
	this.TxTime, eof = source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
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