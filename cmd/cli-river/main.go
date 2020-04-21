package main

import (
	"encoding/json"
	"git.ronaksoftware.com/ronak/riversdk"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/abiosoft/ishell.v2"
	"io/ioutil"
	"os"
)

var (
	_Shell                   *ishell.Shell
	_SDK                     *riversdk.River
	_Log                     *zap.Logger
	_LogLevel                zap.AtomicLevel
	green, red, yellow, blue func(format string, a ...interface{}) string
)

func main() {
	_LogLevel = zap.NewAtomicLevelAt(zap.WarnLevel)
	cfg := zap.NewProductionConfig()
	cfg.Level = _LogLevel
	_Log, _ = cfg.Build()

	green = color.New(color.FgHiGreen).SprintfFunc()
	red = color.New(color.FgHiRed).SprintfFunc()
	yellow = color.New(color.FgHiYellow).SprintfFunc()
	blue = color.New(color.FgHiBlue).SprintfFunc()

	// Initialize Shell
	_Shell = ishell.New()
	_Shell.Println("============================")
	_Shell.Println("## River CLI Console ##")
	_Shell.Println("============================")
	_Shell.AddCmd(Init)
	_Shell.AddCmd(Auth)
	_Shell.AddCmd(Message)
	_Shell.AddCmd(Contact)
	_Shell.AddCmd(SDK)
	_Shell.AddCmd(User)
	_Shell.AddCmd(Debug)
	_Shell.AddCmd(Account)
	_Shell.AddCmd(Group)
	_Shell.AddCmd(File)
	_Shell.AddCmd(Botfather)

	_Shell.Print("River Host (default: river.im):")
	_Shell.Print("DB Path (./_db): ")

	dbPath := _Shell.ReadLine()
	_Shell.Print("DB ID: ")
	dbID := _Shell.ReadLine()
	if dbPath == "" {
		dbPath = "./_db"
	}

	conInfo := new(riversdk.RiverConnection)

	file, err := os.Open("./_connection/connInfo")
	if err == nil {
		b, _ := ioutil.ReadAll(file)
		err := json.Unmarshal(b, conInfo)
		if err != nil {
			_Shell.Print(err.Error())
		}
	}

	conInfo.Delegate = new(ConnInfoDelegates)
	skBytes, _ := ioutil.ReadFile("./keys.json")

	serverEndPoint := "ws://river.ronaksoftware.com"
	fileEndPoint := "http://river.ronaksoftware.com:8080"
	switch len(os.Args) {
	case 3:
		fileEndPoint = os.Args[2]
		fallthrough
	case 2:
		serverEndPoint = os.Args[1]
	}
	_SDK = new(riversdk.River)
	_SDK.SetConfig(&riversdk.RiverConfig{
		ServerEndpoint:         serverEndPoint,
		FileServerEndpoint:     fileEndPoint,
		DbPath:                 dbPath,
		DbID:                   dbID,
		ServerKeys:             ronak.ByteToStr(skBytes),
		MainDelegate:           new(MainDelegate),
		FileDelegate:           new(FileDelegate),
		LogLevel:               int(zapcore.DebugLevel),
		DocumentAudioDirectory: "./_files/audio",
		DocumentVideoDirectory: "./_files/video",
		DocumentPhotoDirectory: "./_files/photo",
		DocumentFileDirectory:  "./_files/file",
		DocumentCacheDirectory: "./_files/cache",
		DocumentLogDirectory:   "./_files/logs",
		ConnInfo:               conInfo,
	})

	// _SDK.TurnOnLiveLogger("http://localhost:2374")
	err = _SDK.Start()
	if err != nil {
		_Log.Fatal(err.Error())
	}
	_SDK.StartNetwork("")
	if _SDK.ConnInfo.AuthID == 0 {
		if err := _SDK.CreateAuthKey(); err != nil {
			_Shell.Println("CreateAuthKey::", err.Error())
		}
	}

	_Shell.Run()

}
