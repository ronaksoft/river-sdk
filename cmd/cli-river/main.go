package main

import (
	"encoding/json"
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/abiosoft/ishell.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	_Shell                   *ishell.Shell
	_SDK                     *riversdk.River
	_Log                     *zap.Logger
	_LogLevel                zap.AtomicLevel
	green, red, yellow, blue func(format string, a ...interface{}) string
)

func main() {
	_LogLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	cfg := zap.NewDevelopmentConfig()
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

	connInfo := new(riversdk.RiverConnection)
	connInfoPath := filepath.Join(dbPath, fmt.Sprintf("connInfo.%s", dbID))
	file, err := os.Open(connInfoPath)
	if err == nil {
		b, _ := ioutil.ReadAll(file)
		err := json.Unmarshal(b, connInfo)
		if err != nil {
			_Shell.Print(err.Error())
		}
	}

	connInfo.Delegate = &ConnInfoDelegates{
		dbPath: dbPath,
		filePath: connInfoPath,
	}



	serverEndPoint := "ws://river.ronaksoftware.com"
	fileEndPoint := "http://river.ronaksoftware.com:8080"
	keysFile := "./keys-staging.json"

	switch len(os.Args) {
	case 3:
		serverEndPoint = os.Args[1]
		fileEndPoint = os.Args[2]
	case 2:
		switch strings.ToLower(os.Args[1]) {
		case "production":
			serverEndPoint = "ws://cyrus.river.im"
			fileEndPoint = "http://file.river.im"
			keysFile = "./keys-production.json"
		case "staging":
			serverEndPoint = "ws://river.ronaksoftware.com"
			fileEndPoint = "http://river.ronaksoftware.com:8080"
			keysFile = "./keys-staging.json"
		}
	}

	skBytes, _ := ioutil.ReadFile(keysFile)

	_SDK = new(riversdk.River)
	_SDK.SetConfig(&riversdk.RiverConfig{
		ServerEndpoint:         serverEndPoint,
		FileServerEndpoint:     fileEndPoint,
		DbPath:                 dbPath,
		DbID:                   dbID,
		ServerKeys:             domain.ByteToStr(skBytes),
		MainDelegate:           new(MainDelegate),
		FileDelegate:           new(FileDelegate),
		LogLevel:               int(zapcore.InfoLevel),
		DocumentAudioDirectory: "./_files/audio",
		DocumentVideoDirectory: "./_files/video",
		DocumentPhotoDirectory: "./_files/photo",
		DocumentFileDirectory:  "./_files/file",
		DocumentCacheDirectory: "./_files/cache",
		DocumentLogDirectory:   "./_files/logs",
		ConnInfo:               connInfo,
	})

	err = _SDK.AppStart()
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
