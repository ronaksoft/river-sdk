package main

import (
	"fmt"
	"os"

	"git.ronaksoftware.com/ronak/riversdk/msg"

	"git.ronaksoftware.com/ronak/riversdk"
	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/abiosoft/ishell.v2"
)

var (
	_DbgPeerID     int64  = 1602004544771208
	_DbgAccessHash uint64 = 4503548912377862
)

var (
	_Shell                                 *ishell.Shell
	_SDK                                   *riversdk.River
	_Log                                   *zap.Logger
	_BLUE, _GREEN, _MAGNETA, _RED, _Yellow func(format string, a ...interface{}) string
)

func main() {
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

	_Shell.Print("River Host (default: river.im):")
	_Shell.Print("DB Path (./_db): ")

	isDebug := os.Getenv("SDK_DEBUG")
	dbPath := ""
	dbID := ""
	if isDebug != "true" {

		dbPath = _Shell.ReadLine()
		_Shell.Print("DB ID: ")

		dbID = _Shell.ReadLine()
		if dbPath == "" {
			dbPath = "./_db"
		}
	} else {
		dbPath = "./_db"
		dbID = "23740071"
	}

	qPath := "./_queue"
	_SDK = new(riversdk.River)
	_SDK.SetConfig(&riversdk.RiverConfig{
		ServerEndpoint:     "ws://new.river.im", //"ws://192.168.1.110/",
		DbPath:             dbPath,
		DbID:               dbID,
		QueuePath:          fmt.Sprintf("%s/%s", qPath, dbID),
		ServerKeysFilePath: "./keys.json",
		MainDelegate:       new(MainDelegate),
		Logger:             new(Logger),
		LogLevel:           int(zapcore.DebugLevel),
	})

	_SDK.Start()
	if _SDK.ConnInfo.AuthID == 0 {
		if err := _SDK.CreateAuthKey(); err != nil {
			_Shell.Println("CreateAuthKey::", err.Error())
		}
	}

	if isDebug != "true" {
		_Shell.Run()
	} else {

		fnRunDebug()

		//block forever
		select {}
	}

}

func fnRunDebug() {
	req := new(msg.ClientSendMessageMedia)
	req.Attributes = make([]*msg.DocumentAttribute, 0)
	req.Caption = "Test file"
	req.ClearDraft = true
	req.FileMIME = ""
	req.FileName = "test.zip"
	req.FilePath = "/home/q/test.zip"
	req.MediaType = msg.InputMediaTypeUploadedDocument
	req.Peer = &msg.InputPeer{
		AccessHash: 4500871196408867,
		ID:         1408226742326241,
		Type:       msg.PeerUser,
	}
	req.ReplyTo = 0
	req.ThumbFilePath = ""
	req.ThumbMIME = ""

	docAttrib := new(msg.DocumentAttribute)
	attrib := new(msg.DocumentAttributeFile)
	attrib.Filename = "test.zip"
	docAttrib.Type = msg.AttributeTypeFilename
	docAttrib.Data, _ = attrib.Marshal()

	req.Attributes = append(req.Attributes, docAttrib)

	buff, _ := req.Marshal()
	reqDelegate := new(RequestDelegate)
	reqID, err := _SDK.ExecuteCommand(msg.C_ClientSendMessageMedia, buff, reqDelegate, false, false)

	_Shell.Println("RequestID :", reqID, "\tError :", err)
}
