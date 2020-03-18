package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/ontio/bonus/config"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
)

const config_file = "./config.json"

func SetOntologyConfig(ctx *cli.Context) error {
	file, err := os.Open(config_file)
	if err != nil {
		return err
	}
	bs, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	cfg := &config.Config{}
	err = json.Unmarshal(bs, cfg)
	if err != nil {
		return err
	}
	config.DefConfig = cfg
	port := ctx.Uint(GetFlagName(RestPortFlag))
	if port != 0 {
		config.DefConfig.RestPort = port
	}
	if config.DefConfig.OntCfg.OntJsonRpcAddressTestNet == "" || config.DefConfig.OntCfg.OntJsonRpcAddressMainNet == "" ||
		config.DefConfig.EthCfg.RpcAddrTestNet == "" || config.DefConfig.EthCfg.RpcAddrMainNet == "" {
		return fmt.Errorf("OntJsonRpcAddress and RpcAddr is nil")
	}
	return nil
}

func PrintErrorMsg(format string, a ...interface{}) {
	format = fmt.Sprintf("\033[31m[ERROR] %s\033[0m\n", format) //Print error msg with red color
	fmt.Printf(format, a...)
}
