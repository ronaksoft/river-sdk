package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/actor"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/controller"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/logs"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/report"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/scenario"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/shared"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/supernumerary"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"gopkg.in/abiosoft/ishell.v2"
	// _ "net/http/pprof"
)

var (
	_Log      *zap.Logger
	_LogLevel zap.AtomicLevel
	_Shell    *ishell.Shell
	_Reporter shared.Reporter
)

func init() {
	_LogLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	cfg := zap.NewProductionConfig()
	cfg.Level = _LogLevel
	_Log, _ = cfg.Build()

	// Initialize Shell
	_Shell = ishell.New()
	_Shell.Println("===============================")
	_Shell.Println("## River Load Tester Console ##")
	_Shell.Println("===============================")

	_Shell.AddCmd(CLI)
	_Shell.AddCmd(cmdRegister)
	_Shell.AddCmd(cmdLogin)
	_Shell.AddCmd(cmdImportContact)
	_Shell.AddCmd(cmdSendMessage)
	_Shell.AddCmd(cmdCreateAuthKey)
	_Shell.AddCmd(cmdClient)
	_Shell.AddCmd(cmdSendFile)
	_Shell.AddCmd(cmdSupernumerary)
	_Shell.AddCmd(cmdDebug)

	logs.SetLogLevel(0) // DBG: -1, INF: 0, WRN: 1, ERR: 2

	_Reporter = report.NewReport()
	// _Reporter.SetIsActive(true)

	if _, err := os.Stat("_cache/"); os.IsNotExist(err) {
		os.Mkdir("_cache/", os.ModePerm)
	}

	loadCachedActors()
}

func main() {

	// init metrics
	boundleID := os.Getenv("CFG_BOUNDLE_ID")
	instanceID := os.Getenv("CFG_INSTANCE_ID")
	shared.InitMetrics(boundleID, instanceID)
	// Run metrics
	go shared.Metrics.Run(2374)

	// // pprof
	// go func() {
	// 	http.ListenAndServe("localhost:6060", nil)
	// }()

	isDebug := os.Getenv("SDK_DEBUG")
	if isDebug == "true" {

		// fnSendContactImport()

		// fnDebugDecrypt()

		// fnSendRawDump()

		// fnPcapParser()

		fnSupernumerary()
	}
	_Shell.Run()

}

func fnSupernumerary() {

	logs.Info("Disabling packet logger ...")
	controller.StopLogginPackets()
	logs.Info("Disabling packet logger ... Done")

	logs.Info("Initializing ...")
	s, err := supernumerary.NewSupernumerary(0, 1000)
	logs.Info("Initializing ... Done")

	if err != nil {
		panic(err)
	}

	// s.CreateAuthKey()
	// s.Register()
	// s.Login()

	logs.Info("SetTickerApplier ...")
	s.SetTickerApplier(30*time.Second, supernumerary.TickerActionSendMessage)
	logs.Info("SetTickerApplier ... Done")

}

func loadCachedActors() {

	fmt.Printf("\n\n Start Loading Cached Actors ... \n\n")

	files, err := ioutil.ReadDir("_cache/")
	if err != nil {
		logs.Error("Fialed to load cached actors LoadCachedActors()", zap.Error(err))
		return
	}

	counter := 0
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		jsonBytes, err := ioutil.ReadFile("_cache/" + f.Name())
		if err != nil {
			fmt.Println("Failed to load actor Filename : ", f.Name())
			continue
		}
		act := new(actor.Actor)
		err = json.Unmarshal(jsonBytes, act)
		if err == nil {
			shared.CacheActor(act)
			counter++
		} else {
			fmt.Println("Failed to Unmarshal actor Filename :", f.Name())
		}
	}
	fmt.Printf("\n Successfully loaded %d actors \n\n", counter)
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
