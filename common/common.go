package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ontio/bonus/config"
)

func IsHave(allStr []string, item string) bool {
	for _, ind := range allStr {
		if ind == item {
			return true
		}
	}
	return false
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func CheckPath(path string) error {
	if PathExists(path) {
		return nil
	}

	dirPath := filepath.Dir(path)
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return fmt.Errorf("mkdir failed: %s", err)
	}

	return nil
}

func (param *ExcelParam) ResetTransferListID() {
	for id, b := range param.BillList {
		b.Id = id
	}
}

func IsTokenTypeSupported(token string) bool {
	for _, t := range config.SupportedTokenTypes {
		if token == t {
			return true
		}
	}
	return false
}
