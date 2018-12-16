package main

import (
	"os"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

var (
	_Shell                                 *ishell.Shell
	_Log                                   *zap.Logger
	_BLUE, _GREEN, _MAGNETA, _RED, _Yellow func(format string, a ...interface{}) string
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

	_BLUE = color.New(color.FgHiBlue).SprintfFunc()
	_GREEN = color.New(color.FgHiGreen).SprintfFunc()
	_MAGNETA = color.New(color.FgHiMagenta).SprintfFunc()
	_RED = color.New(color.FgHiRed).SprintfFunc()
	_Yellow = color.New(color.FgHiYellow).SprintfFunc()

	// Initialize Shell
	_Shell = ishell.New()
	_Shell.Println("============================")
	_Shell.Println("## River Load Tester Console ##")
	_Shell.Println("============================")
	// _Shell.AddCmd(Init)
	// _Shell.AddCmd(Auth)
	// _Shell.AddCmd(Message)
	// _Shell.AddCmd(Contact)
	// _Shell.AddCmd(SDK)
	// _Shell.AddCmd(User)
	// _Shell.AddCmd(Debug)
	// _Shell.AddCmd(Account)
	// _Shell.AddCmd(Tests)
	// _Shell.AddCmd(Group)
}

func main() {
	_Shell.Run()
}
