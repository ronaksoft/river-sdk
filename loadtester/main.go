package main

import (
	"fmt"
	"os"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/actor"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/scenario"

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
	_Shell.Println("===============================")
	_Shell.Println("## River Load Tester Console ##")
	_Shell.Println("===============================")
	_Shell.AddCmd(cmdCreateAuthKey)
	_Shell.AddCmd(cmdRegister)
	_Shell.AddCmd(cmdLogin)
}

func main() {

	act, err := actor.NewActor("237400" + "0000001")
	if err != nil {
		panic(err)
	}

	sCreateAuthKey := scenario.NewCreateAuthKey()
	sCreateAuthKey.Execute(act)
	sCreateAuthKey.Wait()

	// sRegister := scenario.NewRegister()
	// sRegister.Execute(act)
	// sRegister.Wait()

	sLogin := scenario.NewLogin()
	sLogin.Execute(act)
	sLogin.Wait()

	act.PhoneList = []string{"23740071", "23740072"}
	sImportContact := scenario.NewImportContact()
	sImportContact.Execute(act)
	sImportContact.Wait()

	for _, p := range act.Peers {
		fmt.Println(p)
	}

	// TODO : Test SendMessage scenario

	// _Shell.Run()

}
