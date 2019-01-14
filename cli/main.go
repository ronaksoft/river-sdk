package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"

	"git.ronaksoftware.com/ronak/riversdk/filemanager"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/repo"

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
	_Shell.AddCmd(File)

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
		dbID = "23740009"
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
		// fnDecryptDump()
		fnRunUploadFile()
		// fnSendMessageMedia()
		// fnRunDownloadFile()
		// fnSendInputMediaDocument()
		// fnDecodeUpdateHexString()

		//block forever
		select {}
	}

}

func fnSendMessageMedia() {

	dtoFS := repo.Ctx().Files.GetFirstFileStatu()
	fs := filemanager.FileStatus{}
	fs.LoadDTO(dtoFS, nil)
	req := fs.UploadRequest

	// Create SendMessageMedia Request
	x := new(msg.MessagesSendMedia)
	x.Peer = req.Peer
	x.ClearDraft = req.ClearDraft
	x.MediaType = req.MediaType
	x.RandomID = domain.SequentialUniqueID()
	x.ReplyTo = req.ReplyTo

	doc := new(msg.InputMediaUploadedDocument)
	doc.MimeType = req.FileMIME
	doc.Attributes = req.Attributes
	doc.Caption = req.Caption
	doc.File = &msg.InputFile{
		FileID:      fs.FileID,
		FileName:    req.FileName,
		MD5Checksum: "",
		TotalParts:  fs.TotalParts,
	}
	x.MediaData, _ = doc.Marshal()

	reqBuff, _ := x.Marshal()
	reqDelegate := new(RequestDelegate)

	_SDK.ExecuteCommand(msg.C_MessagesSendMedia, reqBuff, reqDelegate, false, false)

}

func fnDecryptDump() {
	file, _ := os.Open("dump.raw")
	rawBytes, _ := ioutil.ReadAll(file)

	protMsg := new(msg.ProtoMessage)
	protMsg.Unmarshal(rawBytes)

	decryptedBytes, _ := domain.Decrypt(_SDK.ConnInfo.AuthKey[:], protMsg.MessageKey, protMsg.Payload)
	encryptedPayload := new(msg.ProtoEncryptedPayload)
	_ = encryptedPayload.Unmarshal(decryptedBytes)

	fmt.Println(encryptedPayload.Envelope)
}

func fnRunUploadFile() {
	req := new(msg.ClientSendMessageMedia)
	req.Attributes = make([]*msg.DocumentAttribute, 0)
	req.Caption = "Test file"
	req.ClearDraft = true
	req.FileMIME = ""
	req.FileName = "test.zip"
	req.FilePath = "/home/q/d.zip"
	req.MediaType = msg.InputMediaTypeUploadedDocument
	// 0009
	req.Peer = &msg.InputPeer{
		AccessHash: 4500232805839723,
		ID:         189353777894340,
		Type:       msg.PeerUser,
	}

	// // 0056
	// req.Peer = &msg.InputPeer{
	// 	AccessHash: 4500871196408867,
	// 	ID:         1408226742326241,
	// 	Type:       msg.PeerUser,
	// }

	// // ZzzzzzzzzzzzzzzzzzzzZ
	// req.Peer = &msg.InputPeer{
	// 	AccessHash: 0,
	// 	ID:         -119344555506722,
	// 	Type:       msg.PeerGroup,
	// }
	req.ReplyTo = 3400
	req.ThumbFilePath = ""
	req.ThumbMIME = ""

	docAttrib := new(msg.DocumentAttribute)
	attrib := new(msg.DocumentAttributeFile)
	attrib.Filename = "test.zip"
	docAttrib.Type = msg.AttributeTypeFile
	docAttrib.Data, _ = attrib.Marshal()

	req.Attributes = append(req.Attributes, docAttrib)

	buff, _ := req.Marshal()
	reqDelegate := new(RequestDelegate)
	reqID, err := _SDK.ExecuteCommand(msg.C_ClientSendMessageMedia, buff, reqDelegate, false, false)

	_Shell.Println("RequestID :", reqID, "\tError :", err)
}

func fnRunDownloadFile() {
	_SDK.FileDownload(7)
}

func fnSendInputMediaDocument() {

	req := new(msg.MessagesSendMedia)
	req.ClearDraft = true
	req.MediaType = msg.InputMediaTypeDocument
	doc := new(msg.InputMediaDocument)
	doc.Caption = "TEST SEND InputMediaDocument"
	doc.Document = new(msg.InputDocument)
	doc.Document.AccessHash = uint64(4499372557768840)
	doc.Document.ClusterID = 1
	doc.Document.ID = int64(2148252319320369373)
	req.MediaData, _ = doc.Marshal()
	// 0056
	req.Peer = &msg.InputPeer{
		AccessHash: 4500871196408867,
		ID:         1408226742326241,
		Type:       msg.PeerUser,
	}
	req.RandomID = domain.SequentialUniqueID()
	req.ReplyTo = 0

	// send the request to server
	buff, _ := req.Marshal()
	reqDelegate := new(RequestDelegate)
	reqID, err := _SDK.ExecuteCommand(msg.C_MessagesSendMedia, buff, reqDelegate, false, false)

	_Shell.Println("RequestID :", reqID, "\tError :", err)

}

func fnDecodeUpdateHexString() {
	str := "0a35081b10c4a7ace5f5862b180120f2c3f1e1052800300038004000480050005a0060d0bdedc8c296c902680070007800800100980100122408d0bdedc8c296c9021204303030391a0430303039220028003000390000000000000000196b73cc19f0fc0f00a00601a8062c"
	buff, err := hex.DecodeString(str)
	if err != nil {
		panic(err)
	}

	udp := new(msg.UpdateNewMessage)
	err = udp.Unmarshal(buff)
	if err != nil {
		panic(err)
	}

	fmt.Println(udp.Message.MediaType)
	fmt.Println(udp.Message.Media)
}
