package main

import (
	"encoding/json"
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk"
	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/abiosoft/ishell.v2"
	"io/ioutil"
	"os"
)

var (
	_DbgPeerID     int64  = 1602004544771208
	_DbgAccessHash uint64 = 4503548912377862
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
	_Shell.AddCmd(Tests)
	_Shell.AddCmd(Group)
	_Shell.AddCmd(File)

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

	qPath := "./_queue"
	_SDK = new(riversdk.River)
	_SDK.SetConfig(&riversdk.RiverConfig{
		ServerEndpoint:     "ws://192.168.1.113:8080", // "ws://test.river.im", // "ws://192.168.1.110/",
		DbPath:             dbPath,
		DbID:               dbID,
		QueuePath:          fmt.Sprintf("%s/%s", qPath, dbID),
		ServerKeysFilePath: "./keys.json",
		MainDelegate:       new(MainDelegate),
		// Logger:                 new(PrintDelegate),
		LogLevel:               int(zapcore.DebugLevel),
		DocumentAudioDirectory: "./_files/audio",
		DocumentVideoDirectory: "./_files/video",
		DocumentPhotoDirectory: "./_files/photo",
		DocumentFileDirectory:  "./_files/file",
		DocumentCacheDirectory: "./_files/cache",
		DocumentLogDirectory:   "./_files/logs",
		ConnInfo:               conInfo,
	})

	_SDK.Start()
	if _SDK.ConnInfo.AuthID == 0 {
		if err := _SDK.CreateAuthKey(); err != nil {
			_Shell.Println("CreateAuthKey::", err.Error())
		}
	}

	_Shell.Run()

}
