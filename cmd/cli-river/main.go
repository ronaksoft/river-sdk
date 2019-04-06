package main

import (
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk"
	"github.com/fatih/color"
	"go.uber.org/zap/zapcore"
	"gopkg.in/abiosoft/ishell.v2"
)

var (
	_DbgPeerID     int64  = 1602004544771208
	_DbgAccessHash uint64 = 4503548912377862
)

var (
	_Shell                   *ishell.Shell
	_SDK                     *riversdk.River
	green, red, yellow, blue func(format string, a ...interface{}) string
)

func main() {

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

	qPath := "./_queue"
	_SDK = new(riversdk.River)
	_SDK.SetConfig(&riversdk.RiverConfig{
		ServerEndpoint:         "ws://new.river.im", // "ws://192.168.1.110/",
		DbPath:                 dbPath,
		DbID:                   dbID,
		QueuePath:              fmt.Sprintf("%s/%s", qPath, dbID),
		ServerKeysFilePath:     "./keys.json",
		MainDelegate:           new(MainDelegate),
		Logger:                 new(PrintDelegate),
		LogLevel:               int(zapcore.DebugLevel),
		DocumentAudioDirectory: "./_files/audio",
		DocumentVideoDirectory: "./_files/video",
		DocumentPhotoDirectory: "./_files/photo",
		DocumentFileDirectory:  "./_files/file",
		DocumentCacheDirectory: "./_files/cache",
		DocumentLogDirectory:   "./_files/logs",
	})

	_SDK.Start()
	if _SDK.ConnInfo.AuthID == 0 {
		if err := _SDK.CreateAuthKey(); err != nil {
			_Shell.Println("CreateAuthKey::", err.Error())
		}
	}

	_Shell.Run()

}