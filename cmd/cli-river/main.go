package main

import (
    "encoding/json"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"

    "github.com/abiosoft/ishell/v2"
    "github.com/fatih/color"
    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/logs"
    "github.com/ronaksoft/river-sdk/sdk/mini"
    riversdk "github.com/ronaksoft/river-sdk/sdk/prime"
    "go.uber.org/zap/zapcore"
)

var (
    _Shell                   *ishell.Shell
    _SDK                     *riversdk.River
    _MiniSDK                 *mini.River
    _DbID                    string
    _DbPath                  string
    _Log                     *logs.Logger
    green, red, yellow, blue func(format string, a ...interface{}) string
)

func main() {
    green = color.New(color.FgHiGreen).SprintfFunc()
    red = color.New(color.FgHiRed).SprintfFunc()
    yellow = color.New(color.FgHiYellow).SprintfFunc()
    blue = color.New(color.FgHiBlue).SprintfFunc()

    _Log = logs.With("RiverCLI")
    // Initialize Shell
    _Shell = ishell.New()

    _Shell.Println("============================")
    _Shell.Println("## River CLI Console ##")
    _Shell.Println("============================")
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
        b, _ := io.ReadAll(file)
        err := json.Unmarshal(b, connInfo)
        if err != nil {
            _Shell.Print(err.Error())
        }
    }

    connInfo.Delegate = &ConnInfoDelegates{
        dbPath:   _DbPath,
        filePath: connInfoPath,
    }

    var serverHostPorts string
    sdkMode := "prime"

    switch len(os.Args) {
    case 2:
        switch strings.ToLower(os.Args[1]) {
        case "production":
            serverHostPorts = "edge.river.im, edge.rivermsg.com"
        case "staging":
            serverHostPorts = "river.ronaksoftware.com"
        case "local":
            serverHostPorts = "localhost"
        case "local2":
            serverHostPorts = "localhost:81"
        default:
            serverHostPorts = os.Args[1]
        }
    case 3:
        switch strings.ToLower(os.Args[1]) {
        case "mini":
            sdkMode = "mini"
        default:
        }
        switch strings.ToLower(os.Args[2]) {
        case "production":
            serverHostPorts = "edge.river.im,edge.rivermsg.com"
        case "staging":
            serverHostPorts = "river.ronaksoftware.com"
        case "local":
            serverHostPorts = "localhost"
        case "local2":
            serverHostPorts = "localhost:81"
        default:
            serverHostPorts = os.Args[1]
        }
    }

    switch sdkMode {
    case "mini":
        _MiniSDK = &mini.River{}
        _MiniSDK.SetConfig(&mini.RiverConfig{
            SeedHostPorts:          serverHostPorts,
            DbPath:                 _DbPath,
            DbID:                   _DbID,
            MainDelegate:           new(MainDelegate),
            LogLevel:               int(zapcore.DebugLevel),
            DocumentAudioDirectory: "./_files/audio",
            DocumentVideoDirectory: "./_files/video",
            DocumentPhotoDirectory: "./_files/photo",
            DocumentFileDirectory:  "./_files/file",
            DocumentCacheDirectory: "./_files/cache",
            LogDirectory:           "./_files/logs",
            ConnInfo: &mini.RiverConnection{
                AuthID:  connInfo.AuthID,
                AuthKey: connInfo.AuthKey[:],
                UserID:  connInfo.UserID,
            },
        })

        err = _MiniSDK.AppStart()
        if err != nil {
            panic(err)
        }

        loadCommands(
            Mini,
        )

    default:
        _SDK = &riversdk.River{}
        _SDK.SetConfig(&riversdk.RiverConfig{
            SeedHostPorts:          serverHostPorts,
            DbPath:                 _DbPath,
            DbID:                   _DbID,
            MainDelegate:           new(MainDelegate),
            FileDelegate:           new(FileDelegate),
            CallDelegate:           new(CallDelegate),
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
                Limit:  500,
            }
            reqBytes, _ := req.Marshal()
            delegate := new(RequestDelegate)
            _, _ = _SDK.ExecuteCommand(msg.C_MessagesGetDialogs, reqBytes, delegate)
        }
        loadCommands(
            Account, Auth, Bot, Contact, Debug, File, Gif, Group, Init, Label, Message, SDK, System, Team, User, WallPaper, Call,
        )

    }

    _Shell.Run()

}

func loadCommands(cmds ...*ishell.Cmd) {
    for _, cmd := range cmds {
        _Shell.AddCmd(cmd)
    }
}
