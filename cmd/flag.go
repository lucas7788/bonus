package cmd

import (
	"strings"

	"github.com/ontio/bonus/config"
	"github.com/urfave/cli"
)

var (
	LogLevelFlag = cli.UintFlag{
		Name:  "loglevel",
		Usage: "Set the log level to `<level>` (0~6). 0:Trace 1:Debug 2:Info 3:Warn 4:Error 5:Fatal 6:MaxLevel",
		Value: uint(config.DEFAULT_LOG_LEVEL),
	}
	RestPortFlag = cli.UintFlag{
		Name:  "restport",
		Usage: "restful server listening port `<number>`",
		Value: 0,
	}
	ProjectIdFlag = cli.UintFlag{
		Name:  "projectid",
		Usage: "airdrop project id `<number>`",
		Value: 0,
	}
)

func GetFlagName(flag cli.Flag) string {
	name := flag.GetName()
	if name == "" {
		return ""
	}
	return strings.TrimSpace(strings.Split(name, ",")[0])
}
