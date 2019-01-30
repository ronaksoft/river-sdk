package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/actor"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/report"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/scenario"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"
	"git.ronaksoftware.com/ronak/riversdk/log"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	ishell "gopkg.in/abiosoft/ishell.v2"
	// _ "net/http/pprof"
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
	_Shell.AddCmd(cmdCreateAuthKey)

	_Shell.AddCmd(cmdDebug)

	log.SetLogger(Log)
	log.SetLogLevel(-1) // DBG: -1, INF: 0, WRN: 1, ERR: 2

	_Reporter = report.NewReport()
}

func main() {

	// // pprof
	// go func() {
	// 	http.ListenAndServe("localhost:6060", nil)
	// }()

	isDebug := os.Getenv("SDK_DEBUG")
	if isDebug == "true" {

		// fnSendContactImport()

		// fnDebugDecrypt()

		// fnSendRawDump()
	}

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

func fnSendContactImport() {
	act, err := actor.NewActor("2374000009953")
	if err != nil {
		panic(err)
	}
	act.SetPhoneList([]string{"23740072"})
	sn := scenario.NewImportContact(true)
	sn.Play(act)
	sn.Wait(act)
}

func fnSendRawDump() {
	wsDialer := websocket.DefaultDialer
	wsDialer.ReadBufferSize = 32 * 1024  // 32KB
	wsDialer.WriteBufferSize = 32 * 1024 // 32KB
	conn, _, err := wsDialer.Dial(shared.DefaultServerURL, nil)
	if err != nil {
		panic(err)
	}

	buff, err := ioutil.ReadFile("ImportContact_Dump.raw")
	if err != nil {
		panic(err)
	}

	err = conn.WriteMessage(websocket.BinaryMessage, buff)
	if err != nil {
		panic(err)
	}
}

func fnDebugDecrypt() {
	hexStr := "08a4e4f0e081dbd1869601122043cac6c21108542e37ac695a3658c0975fc55fb12f6468d14c765add167869601a43be33958805fde4bb686b58c4566eaee3c1b289fe5ca3d434f41a6fcb51b426430821faf35c50e7aacf46faf3e62ac710c9a2a261a9f6e12b48937c3821c1718b20e5ef"
	rawbytes, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	act, err := actor.NewActor("2374000009953")
	if err != nil {
		panic(err)
	}
	authID, authKey := act.GetAuthInfo()
	fmt.Println(authID)
	decryptProtoMessage(rawbytes, authKey)

}
