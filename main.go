package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/ontio/bonus/cmd"
	"github.com/ontio/bonus/common"
	"github.com/ontio/bonus/config"
	"github.com/ontio/bonus/manager/interfaces"
	"github.com/ontio/bonus/restful"
	"github.com/ontio/ontology/common/log"
	"github.com/urfave/cli"
	"os/exec"
	"time"
)

func setupAPP() *cli.App {
	app := cli.NewApp()
	app.Usage = "Bonus CLI"
	app.Action = startBonus
	app.Version = config.Version
	app.Copyright = "Copyright in 2018 The Ontology Authors"
	app.Flags = []cli.Flag{
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
	log.Infof("bonus version %s", config.Version)
	if !common.PathExists(config.DBPath) {
		err := os.Mkdir(config.DBPath, os.ModePerm)
		if err != nil {
			log.Errorf("Mkdir error: %s", err)
			return
		}
	}

	if err := initConfig(ctx); err != nil {
		log.Errorf("initConfig error: %s", err)
		return
	}
	if err := restful.StartServer(); err != nil {
		log.Errorf("start web server: %s", err)
		return
	}
	startHtml()
	log.Info("startHtml success")
	time.Sleep(1 * time.Second)
	err := openBrowser("http://127.0.0.1:20328")
	if err != nil {
		log.Infof("openBrowser failed: %s", err)
	}
	log.Infof("Please open: %s in browser", "http://127.0.0.1:20328")
	waitToExit()
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
			restful.DefBonusMap.Range(func(key, value interface{}) bool {
				mgr, ok := value.(interfaces.WithdrawManager)
				if ok {
					mgr.CloseDB()
				}
				return true
			})
			log.Infof("bonus server received exit signal: %s.", sig.String())
			close(exit)
			break
		}
	}()
	<-exit
}

func startHtml() {
	r := gin.Default()
	// work version
	r.Static("/", "web")
	r.NoRoute(func(c *gin.Context) {
		log.Info("************")
		c.Redirect(301, "/")
	})
	go func() {
		err := r.Run(":20328")
		if err != nil {
			log.Errorf("startHtml err: %s", err)
		}
		log.Info("web server over")
	}()
}

func openBrowser(uri string) error {
	var commands = map[string]string{
		"windows": "start",
		"darwin":  "open",
		"linux":   "xdg-open",
	}

	run, ok := commands[runtime.GOOS]
	if !ok {
		return fmt.Errorf("don't know how to open things on %s platform", runtime.GOOS)
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command(`cmd`, `/c`, `start`, `http://127.0.0.1:20328`)
	} else {
		cmd = exec.Command(run, uri)
	}
	return cmd.Start()
}
