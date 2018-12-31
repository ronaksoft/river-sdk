package main

import (
	"os"

	"git.ronaksoftware.com/ronak/riversdk/log"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

var (
	_Shell                *ishell.Shell
	_Log                  *zap.Logger
	_GREEN, _RED, _Yellow func(format string, a ...interface{}) string
)

func init() {
	logConfig := zap.NewProductionConfig()
	logConfig.Encoding = "console"
	logConfig.Level = zap.NewAtomicLevelAt(zapcore.Level(zapcore.DebugLevel))
	if v, err := logConfig.Build(); err != nil {
		os.Exit(1)
	} else {
		_Log = v
	}

	_GREEN = color.New(color.FgHiGreen).SprintfFunc()
	_RED = color.New(color.FgHiRed).SprintfFunc()
	_Yellow = color.New(color.FgHiYellow).SprintfFunc()

	// Initialize Shell
	_Shell = ishell.New()
	_Shell.Println("===============================")
	_Shell.Println("## River Load Tester Console ##")
	_Shell.Println("===============================")
	_Shell.AddCmd(cmdRegister)
	_Shell.AddCmd(cmdLogin)
	_Shell.AddCmd(cmdImportContact)
	_Shell.AddCmd(cmdSendMessage)

	log.SetLogger(Log)
}

func main() {

	// act, err := actor.NewActor("237400" + "0000001")
	// act.SetPhoneList([]string{"23740071", "23740072"})
	// if err != nil {
	// 	panic(err)
	// }

	// scenario.Play(act, scenario.NewCreateAuthKey())
	// scenario.Play(act, scenario.NewRegister())
	// scenario.Play(act, scenario.NewLogin())
	// scenario.Play(act, scenario.NewImportContact())
	// scenario.Play(act, scenario.NewSendMessage())

	_Shell.Run()
}

// Log log printer
func Log(logLevel int, msg string) {

	switch logLevel {
	case int(zap.DebugLevel):
		_Shell.Println("DBG : \t", msg)
	case int(zap.WarnLevel):
		_Shell.Println(_Yellow("WRN : \t %s", msg))
	case int(zap.InfoLevel):
		_Shell.Println(_GREEN("INF : \t %s", msg))
	case int(zap.ErrorLevel):
		_Shell.Println(_RED("ERR : \t %s", msg))
	case int(zap.FatalLevel):
		_Shell.Println(_RED("FTL : \t %s", msg))
	}
}
