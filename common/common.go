package common

import (
	"fmt"
	"github.com/ontio/ontology/common/log"
	"os"
	"strings"
	"path/filepath"
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
