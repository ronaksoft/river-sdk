package main

import (
	"encoding/json"
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	riversdk "git.ronaksoft.com/river/sdk/sdk/prime"
	"github.com/fatih/color"
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
	_DbID                    string
	_DbPath                  string
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
	_Shell.AddCmd(Bot)
	_Shell.AddCmd(User)
	_Shell.AddCmd(Debug)
	_Shell.AddCmd(Account)
	_Shell.AddCmd(Group)
	_Shell.AddCmd(File)
	_Shell.AddCmd(Botfather)
	_Shell.AddCmd(WallPaper)
	_Shell.AddCmd(Label)
	_Shell.AddCmd(Gif)
	_Shell.AddCmd(Team)
	_Shell.AddCmd(System)

	_Shell.Print("River Host (default: river.im):")
	_Shell.Print("DB Path (./_db): ")

	_DbPath = _Shell.ReadLine()
	_Shell.Print("DB ID: ")
	_DbID = _Shell.ReadLine()
	if _DbPath == "" {
		_DbPath = "./_db"
	}

	connInfo := new(riversdk.RiverConnection)
	connInfoPath := filepath.Join(_DbPath, fmt.Sprintf("connInfo.%s", _DbID))
	file, err := os.Open(connInfoPath)
	if err == nil {
		b, _ := ioutil.ReadAll(file)
		err := json.Unmarshal(b, connInfo)
		if err != nil {
			_Shell.Print(err.Error())
		}
	}

	connInfo.Delegate = &ConnInfoDelegates{
		dbPath:   _DbPath,
		filePath: connInfoPath,
	}

	serverHostPort := "river.ronaksoftware.com"

	switch len(os.Args) {
	case 2:
		switch strings.ToLower(os.Args[1]) {
		case "production":
			serverHostPort = "edge.river.im"
		case "staging":
			serverHostPort = "river-rony.ronaksoftware.com"
		case "local":
			serverHostPort = "localhost"
		case "local2":
			serverHostPort = "localhost:81"
		default:
			serverHostPort = os.Args[1]
		}
	}

	_SDK = &riversdk.River{}
	_SDK.SetConfig(&riversdk.RiverConfig{
		ServerHostPort:         serverHostPort,
		DbPath:                 _DbPath,
		DbID:                   _DbID,
		MainDelegate:           new(MainDelegate),
		FileDelegate:           new(FileDelegate),
		LogLevel:               int(zapcore.InfoLevel),
		DocumentAudioDirectory: "./_files/audio",
		DocumentVideoDirectory: "./_files/video",
		DocumentPhotoDirectory: "./_files/photo",
		DocumentFileDirectory:  "./_files/file",
		DocumentCacheDirectory: "./_files/cache",
		LogDirectory:           "./_files/logs",
		ConnInfo:               connInfo,
	})

	err = _SDK.AppStart()
	if err != nil {
		panic(err)
	}
	_SDK.StartNetwork("")
	if _SDK.ConnInfo.AuthID == 0 {
		if err := _SDK.CreateAuthKey(); err != nil {
			_Shell.Println("CreateAuthKey::", err.Error())
		}
	}

	if _SDK.ConnInfo.UserID != 0 {
		req := &msg.MessagesGetDialogs{
			Offset: 0,
			Limit: 500,
		}
		reqBytes, _ := req.Marshal()
		delegate := new(RequestDelegate)
		_, _ = _SDK.ExecuteCommand(msg.C_MessagesGetDialogs, reqBytes, delegate)
	}

	_Shell.Run()

}
