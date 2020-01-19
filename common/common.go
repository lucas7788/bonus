package common

import (
	"strings"
	"fmt"
	"os"
	"github.com/ontio/ontology/common/log"
)


func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func CheckPath(path string) error {
	if !PathExists(path) {
		paths := strings.Split(path, "/")
		var tempDir string
		for i := 0; i < len(paths); i++ {
			if paths[i] == "." {
				tempDir = paths[i]
				continue
			}
			tempDir = fmt.Sprintf("%s%s%s", tempDir, string(os.PathSeparator), paths[i])
			if !PathExists(tempDir) {
				err := os.Mkdir(tempDir, os.ModePerm)
				if err != nil {
					log.Errorf("Mkdir failed, error: %s", err)
					return fmt.Errorf("Mkdir failed, error: %s", err)
				}
			}
		}
	}
	return nil
}
