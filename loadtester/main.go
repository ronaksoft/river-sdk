package main

import (
	"os"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/report"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"
	"git.ronaksoftware.com/ronak/riversdk/log"

	_ "net/http/pprof"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

var (
	_Shell                *ishell.Shell
	_Log                  *zap.Logger
	_GREEN, _RED, _Yellow func(format string, a ...interface{}) string
	_Reporter             shared.Reporter
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

	_Shell.AddCmd(CLI)
	_Shell.AddCmd(cmdPrint)
	_Shell.AddCmd(cmdRegister)
	_Shell.AddCmd(cmdLogin)
	_Shell.AddCmd(cmdImportContact)
	_Shell.AddCmd(cmdSendMessage)

	_Shell.AddCmd(cmdRegisterByPool)
	_Shell.AddCmd(cmdLoginByPool)
	_Shell.AddCmd(cmdImportContactByPool)
	_Shell.AddCmd(cmdSendMessageByPool)

	log.SetLogger(Log)
	log.SetLogLevel(0) // DBG: -1, INF: 1, WRN: 2, ERR: 3

	_Reporter = report.NewReport()
}

func main() {

	// // pprof
	// go func() {
	// 	http.ListenAndServe("localhost:6060", nil)
	// }()

	_Shell.Run()
}

// Log log printer
func Log(logLevel int, msg string) {
	if _Reporter.IsActive() {
		return
	}

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
