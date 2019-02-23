package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"

	"git.ronaksoftware.com/ronak/riversdk/logs"
	"github.com/fatih/color"
	"go.uber.org/zap"

	"git.ronaksoftware.com/ronak/riversdk/filemanager"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/repo"

	"git.ronaksoftware.com/ronak/riversdk/msg"

	"git.ronaksoftware.com/ronak/riversdk"
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
		ServerEndpoint:         "ws://new.river.im", //"ws://192.168.1.110/",
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
		// fnRunUploadFile()
		// fnSendMessageMedia()
		// fnRunDownloadFile()
		// fnSendInputMediaDocument()
		// fnDecodeUpdateHexString()
		// fnMessagesReadContents()
		// fnGetDialogs()
		// fnAccountUploadPhoto()
		// fnGroupUploadPhoto()
		// fnLoginWithAuthKey()
		// fnRunDownloadFileThumbnail()

		//block forever
		select {}
	}

}

func fnLoginWithAuthKey() {
	req := new(msg.AuthLoginByToken)
	req.Token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJleHAiOjUwMDExMTg0MzcsImVtYWlsIjpudWxsLCJ1c2VyX2lkIjoyNCwidXNlcm5hbWUiOiIwOTAxNjg3NjA0MCIsImlwIjoiMi4xNzYuNzQuMjgifQ.ajYyfUPtnCoQMwh6gw0gsuyEAalyD54wtow0JirZkbI"
	req.Provider = "ap"
	// AuthProviderAsanPardakht = "ap"
	// AuthProviderNested       = "nested"
	// AuthProviderGoogle       = "google"
	// AuthProviderApple        = "apple"

	reqBytes, _ := req.Marshal()
	reqDelegate := new(RequestDelegate)
	if _, err := _SDK.ExecuteCommand(msg.C_AuthLoginByToken, reqBytes, reqDelegate, false, false); err != nil {
		logs.Error("ExecuteCommand failed", zap.Error(err))
	}

}

func fnAccountUploadPhoto() {
	_SDK.AccountUploadPhoto("/home/q/Desktop/decrypt_dump.raw.png")
}

func fnGetDialogs() {
	req := msg.MessagesGetDialogs{}
	req.Limit = int32(100)
	req.Offset = int32(0)

	reqBytes, _ := req.Marshal()
	reqDelegate := new(RequestDelegate)
	if _, err := _SDK.ExecuteCommand(msg.C_MessagesGetDialogs, reqBytes, reqDelegate, false, false); err != nil {
		logs.Error("ExecuteCommand failed", zap.Error(err))
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
		TotalParts:  int32(fs.TotalParts),
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
	req.FilePath = "_f/video.mp4"
	req.ThumbFilePath = "/tmpfs/thumb.jpg"
	req.MediaType = msg.InputMediaTypeUploadedDocument
	// 0009
	req.Peer = &msg.InputPeer{
		AccessHash: 4501753917828959,
		ID:         986281403829488,
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
	req.ThumbMIME = "image/jpg"

	docAttrib := new(msg.DocumentAttribute)
	attrib := new(msg.DocumentAttributeFile)
	attrib.Filename = "test.png"
	docAttrib.Type = msg.AttributeTypeFile
	docAttrib.Data, _ = attrib.Marshal()

	req.Attributes = append(req.Attributes, docAttrib)

	buff, _ := req.Marshal()
	reqDelegate := new(RequestDelegate)
	reqID, err := _SDK.ExecuteCommand(msg.C_ClientSendMessageMedia, buff, reqDelegate, false, false)

	_Shell.Println("RequestID :", reqID, "\tError :", err)
}

func fnRunDownloadFile() {

	// oneAndHalf := domain.FilePayloadSize + (domain.FilePayloadSize / 2)
	// twoAndHalf := oneAndHalf + domain.FilePayloadSize
	// buff := make([]byte, oneAndHalf)
	// rand.Read(buff)
	// ioutil.WriteFile("1.5.raw", buff, os.ModePerm)

	// buff = make([]byte, twoAndHalf)
	// rand.Read(buff)
	// ioutil.WriteFile("2.5.raw", buff, os.ModePerm)

	// threeAndHalf := (domain.FilePayloadSize * 3) + (domain.FilePayloadSize / 2)
	// buff := make([]byte, threeAndHalf)
	// rand.Read(buff)
	// ioutil.WriteFile("3.5.raw", buff, os.ModePerm)

	_SDK.FileDownload(6)
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
	str := ""
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

func fnMessagesReadContents() {

	req := new(msg.MessagesReadContents)
	// 0056
	req.Peer = &msg.InputPeer{
		AccessHash: 4500871196408867,
		ID:         1408226742326241,
		Type:       msg.PeerUser,
	}
	req.MessageIDs = []int64{
		4038,
	}

	// send the request to server
	buff, _ := req.Marshal()
	reqDelegate := new(RequestDelegate)
	reqID, err := _SDK.ExecuteCommand(msg.C_MessagesReadContents, buff, reqDelegate, false, false)

	_Shell.Println("RequestID :", reqID, "\tError :", err)

}

func fnGroupUploadPhoto() {
	_SDK.GroupUploadPhoto(-2101046409375509, "/home/q/Desktop/decrypt_dump.raw.png")
}

func fnRunDownloadFileThumbnail() {
	_SDK.FileDownloadThumbnail(662)
}
