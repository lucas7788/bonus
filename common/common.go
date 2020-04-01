package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ontio/bonus/config"
	"github.com/ontio/ontology/common/log"
	"io/ioutil"
	"sort"
	"strings"
)

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

func (param *ExcelParam) TrParamSort() {
	sort.Slice(param.BillList, func(i, j int) bool {
		if param.BillList[i].Id < param.BillList[j].Id {
			return true
		}
		return false
	})
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

func getBaseDir() string {
	return filepath.Join(".", config.DBPath)
}

func GetEventDir(tokenType string, eventType string) string {
	return filepath.Join(getBaseDir(), tokenType+"_"+eventType)
}

func GetEventDBFilename(net, tokenType, eventType string) string {
	return filepath.Join(GetEventDir(tokenType, eventType), net, net+".db")
}

func ClearData(netty, tokenty, evtty string) {
	dbFileName := GetEventDBFilename(netty, tokenty, evtty)
	if PathExists(dbFileName) {
		err := os.Remove(dbFileName)
		if err != nil {
			log.Errorf("[clearData] Remove dbFileName error:%s", err)
		}
	}
	evtPath := GetEventDir(tokenty, evtty)
	if PathExists(evtPath) {
		log.Infof("removeall dir:%s", evtPath)
		err := os.RemoveAll(evtPath)
		if err != nil {
			log.Errorf("[clearData] Remove evtdir error:%s", err)
		}
	}
}

//
// return { "ONT/ONG/OEP4/ETH/ERC20" + event-name }
//
func GetAllEventDirs() ([]string, error) {
	files, err := ioutil.ReadDir(getBaseDir())
	if err != nil {
		return nil, fmt.Errorf("failed to read basedir: %s", err)
	}
	events := make([]string, 0)
	for _, f := range files {
		if f.IsDir() {
			eventName := f.Name()
			for _, tokenName := range config.SupportedTokenTypes {
				if strings.HasPrefix(eventName, tokenName+"_") {
					events = append(events, eventName)
					break
				}
			}
		}
	}
	return events, nil
}
