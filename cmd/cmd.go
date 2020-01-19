package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/CandyDrop/cmd/utils"
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
	port := ctx.Uint(utils.GetFlagName(utils.RestPortFlag))
	if port != 0 {
		config.DefConfig.RestPort = port
	}
	return nil
}

func PrintErrorMsg(format string, a ...interface{}) {
	format = fmt.Sprintf("\033[31m[ERROR] %s\033[0m\n", format) //Print error msg with red color
	fmt.Printf(format, a...)
}
