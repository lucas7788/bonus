package main

import (
	"fmt"
	"github.com/ontio/bonus/cmd"
	"github.com/ontio/bonus/config"
	"github.com/ontio/bonus/db"
	"github.com/ontio/bonus/restful"
	"github.com/ontio/ontology/common/log"
	"github.com/urfave/cli"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func setupAPP() *cli.App {
	app := cli.NewApp()
	app.Usage = "Bonus CLI"
	app.Action = startBonus
	app.Version = config.Version
	app.Copyright = "Copyright in 2018 The Ontology Authors"
	app.Flags = []cli.Flag{
		//common setting
		cmd.LogLevelFlag,
		cmd.RestPortFlag,
	}
	app.Before = func(context *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		return nil
	}
	return app
}

func main() {
	if err := setupAPP().Run(os.Args); err != nil {
		cmd.PrintErrorMsg(err.Error())
		os.Exit(1)
	}
}

func startBonus(ctx *cli.Context) {
	initLog(ctx)
	log.Infof("ontology version %s", config.Version)
	err := initConfig(ctx)
	if err != nil {
		log.Errorf("initConfig error: %s", err)
		return
	}
	//err = initDB(ctx)
	//if err != nil {
	//	log.Errorf("initDB error: %s", err)
	//	return
	//}
	restful.StartServer()
	waitToExit()
}

func initDB(ctx *cli.Context) error {
	dberr := db.ConnectDB()
	if dberr != nil {
		log.Errorf("username: %s, password: %s", config.DefConfig.ProjectDBUser,
			config.DefConfig.ProjectDBPassword)
		return fmt.Errorf("ConnectDB error: %s", dberr)
	}
	return nil
}

func initLog(ctx *cli.Context) {
	//init log module
	logLevel := ctx.GlobalInt(cmd.GetFlagName(cmd.LogLevelFlag))
	logName := fmt.Sprintf("%s%s", config.LogPath, string(os.PathSeparator))
	log.InitLog(logLevel, logName, log.Stdout)
}

func initConfig(ctx *cli.Context) error {
	//init config
	return cmd.SetOntologyConfig(ctx)
}

func waitToExit() {
	exit := make(chan bool, 0)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for sig := range sc {
			log.Infof("bonus server received exit signal: %s.", sig.String())
			close(exit)
			break
		}
	}()
	<-exit
}
