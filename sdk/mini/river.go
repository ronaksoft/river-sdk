package mini

import (
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	fileCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_file"
	networkCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_network"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/repo"
	riversdk "git.ronaksoft.com/river/sdk/sdk/prime"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/tools"
	"go.uber.org/zap"
	"strings"
	"sync"
	"time"
)

/*
   Creation Time: 2021 - Apr - 28
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func SetLogLevel(l int) {
	logs.SetLogLevel(l)
}

type RiverConfig struct {
	ServerHostPort string
	// DbPath is the path of the folder holding the sqlite database.
	DbPath string
	// DbID is used to save data for different accounts in separate databases. Could be used for multi-account cases.
	DbID string
	// LogLevel
	LogLevel  int
	SentryDSN string
	// Folder path to save files
	DocumentPhotoDirectory string
	DocumentVideoDirectory string
	DocumentFileDirectory  string
	DocumentAudioDirectory string
	DocumentCacheDirectory string
	LogDirectory           string
	// ConnInfo stores the Connection Info
	ConnInfo *RiverConnection
	// ClientInfo
	ClientPlatform string
	ClientVersion  string
	ClientOs       string
	ClientVendor   string
	CountryCode    string

	MaxInFlightDownloads int32
	MaxInFlightUploads   int32

	// Team related parameters
	TeamID         int64
	TeamAccessHash int64
}

// River
// This the main and a wrapper around all the components of the system (networkController, queueController,
// syncController). All the controllers could be used standalone, but this SDK connect them in a way
// we think is the best possible.
// Only the functions which are exposed will be used by the user of the SDK. All the low-level tasks
// to smooth the connection between client and server are done by this SDK. The underlying storage used
// by this SDK is Badger V2. 'repo' is the package name selected to handle repository functions.
type River struct {
	ConnInfo       *RiverConnection
	serverHostPort string
	dbPath         string
	sentryDSN      string

	// localCommands can be satisfied by client cache
	localCommands map[int64]domain.LocalMessageHandler

	// Internal Controllers
	networkCtrl *networkCtrl.Controller
	fileCtrl    *fileCtrl.Controller

	// Delegates
	mainDelegate riversdk.MainDelegate
}

// SetConfig must be called before any other function, otherwise it panics
func (r *River) SetConfig(conf *RiverConfig) {
	domain.ClientPlatform = conf.ClientPlatform
	domain.ClientVersion = conf.ClientVersion
	domain.ClientOS = conf.ClientOs
	domain.ClientVendor = conf.ClientVendor

	r.sentryDSN = conf.SentryDSN
	r.ConnInfo = conf.ConnInfo
	r.serverHostPort = conf.ServerHostPort

	if conf.MaxInFlightDownloads <= 0 {
		conf.MaxInFlightDownloads = 10
	}
	if conf.MaxInFlightUploads <= 0 {
		conf.MaxInFlightUploads = 10
	}

	// Initialize DB Path
	if strings.HasPrefix(conf.DbPath, "file://") {
		conf.DbPath = conf.DbPath[7:]
	}
	conf.DbPath = strings.TrimRight(conf.DbPath, "/ ")
	r.dbPath = fmt.Sprintf("%s/%s.db", conf.DbPath, conf.DbID)

	r.registerCommandHandlers()

	// set log level
	logs.SetLogLevel(conf.LogLevel)

	// set log file path
	if conf.LogDirectory != "" {
		_ = logs.SetLogFilePath(conf.LogDirectory)
	}

	// Initialize Network Controller
	r.networkCtrl = networkCtrl.New(
		networkCtrl.Config{
			WebsocketEndpoint: fmt.Sprintf("ws://%s", conf.ServerHostPort),
			HttpEndpoint:      fmt.Sprintf("http://%s", conf.ServerHostPort),
			CountryCode:       conf.CountryCode,
		},
	)
	r.networkCtrl.OnNetworkStatusChange = func(newQuality domain.NetworkStatus) {}
	r.networkCtrl.OnGeneralError = r.onGeneralError
	r.networkCtrl.OnMessage = r.onReceivedMessage
	r.networkCtrl.OnUpdate = r.onReceivedUpdate
	r.networkCtrl.OnWebsocketConnect = r.onNetworkConnect

	// Initialize FileController
	repo.Files.SetRootFolders(
		conf.DocumentAudioDirectory,
		conf.DocumentFileDirectory,
		conf.DocumentPhotoDirectory,
		conf.DocumentVideoDirectory,
		conf.DocumentCacheDirectory,
	)

	// Initialize River Connection
	logs.Info("River SetConfig done!")

	// Set current team
	domain.SetCurrentTeam(conf.TeamID, uint64(conf.TeamAccessHash))
}

func (r *River) onNetworkConnect() (err error) {
	return nil
}

func (r *River) onGeneralError(requestID uint64, e *rony.Error) {
	logs.Info("We received error (General)",
		zap.Uint64("ReqID", requestID),
		zap.String("Code", e.Code),
		zap.String("Item", e.Items),
	)

	if r.mainDelegate != nil && requestID == 0 {
		buff, _ := e.Marshal()
		r.mainDelegate.OnGeneralError(buff)
	}
}

func (r *River) onReceivedMessage(msgs []*rony.MessageEnvelope) {}

func (r *River) onReceivedUpdate(updateContainer *msg.UpdateContainer) {}

func (r *River) postUploadProcess(uploadRequest *msg.ClientFileRequest) bool {
	defer logs.RecoverPanic(
		"River::postUploadProcess",
		domain.M{
			"OS":       domain.ClientOS,
			"Ver":      domain.ClientVersion,
			"FilePath": uploadRequest.FilePath,
		},
		nil,
	)

	logs.Info("River Post Upload Process",
		zap.Bool("IsProfile", uploadRequest.IsProfilePhoto),
		zap.Int64("MessageID", uploadRequest.MessageID),
		zap.Int64("FileID", uploadRequest.FileID),
	)
	switch {
	case uploadRequest.IsProfilePhoto == false && uploadRequest.MessageID != 0:
		return r.sendMessageMedia(uploadRequest)
	}
	return false
}
func (r *River) sendMessageMedia(uploadRequest *msg.ClientFileRequest) (success bool) {
	// This is a upload for message send
	pendingMessage := repo.PendingMessages.GetByID(uploadRequest.MessageID)
	if pendingMessage == nil {
		return true
	}

	req := &msg.ClientSendMessageMedia{}
	_ = req.Unmarshal(pendingMessage.Media)
	err := tools.Try(3, time.Millisecond*500, func() error {
		var fileLoc *msg.FileLocation
		if uploadRequest.FileID != 0 && uploadRequest.AccessHash != 0 && uploadRequest.ClusterID != 0 {
			req.MediaType = msg.InputMediaType_InputMediaTypeDocument
			fileLoc = &msg.FileLocation{
				ClusterID:  uploadRequest.ClusterID,
				FileID:     uploadRequest.FileID,
				AccessHash: uploadRequest.AccessHash,
			}
		}
		return repo.PendingMessages.UpdateClientMessageMedia(pendingMessage, uploadRequest.TotalParts, req.MediaType, fileLoc)
	})
	if err != nil {
		logs.Error("Error On UpdateClientMessageMedia", zap.Error(err))
	}

	// Create SendMessageMedia Request
	x := &msg.MessagesSendMedia{
		Peer:       req.Peer,
		ClearDraft: req.ClearDraft,
		MediaType:  req.MediaType,
		RandomID:   pendingMessage.FileID,
		ReplyTo:    req.ReplyTo,
	}

	switch x.MediaType {
	case msg.InputMediaType_InputMediaTypeUploadedDocument:
		doc := &msg.InputMediaUploadedDocument{
			MimeType:   req.FileMIME,
			Attributes: req.Attributes,
			Caption:    req.Caption,
			Entities:   req.Entities,
			File: &msg.InputFile{
				FileID:      uploadRequest.FileID,
				FileName:    req.FileName,
				MD5Checksum: "",
			},
			TinyThumbnail: req.TinyThumb,
		}
		if uploadRequest.ThumbID != 0 {
			doc.Thumbnail = &msg.InputFile{
				FileID:      uploadRequest.ThumbID,
				FileName:    "thumb_" + req.FileName,
				MD5Checksum: "",
			}
		}
		x.MediaData, _ = doc.Marshal()
	case msg.InputMediaType_InputMediaTypeDocument:
		doc := &msg.InputMediaDocument{
			Caption:    req.Caption,
			Attributes: req.Attributes,
			Entities:   req.Entities,
			Document: &msg.InputDocument{
				ID:         uploadRequest.FileID,
				AccessHash: uploadRequest.AccessHash,
				ClusterID:  uploadRequest.ClusterID,
			},
			TinyThumbnail: req.TinyThumb,
		}
		if uploadRequest.ThumbID != 0 {
			doc.Thumbnail = &msg.InputFile{
				FileID:      uploadRequest.ThumbID,
				FileName:    "thumb_" + req.FileName,
				MD5Checksum: "",
			}
		}
		x.MediaData, _ = doc.Marshal()

	default:
	}
	// reqBuff, _ := x.Marshal()
	success = true

	waitGroup := sync.WaitGroup{}
	// waitGroup.Add(1)
	// successCB := func(m *rony.MessageEnvelope) {
	// 	logs.Info("MessagesSendMedia success callback called", zap.String("C", registry.ConstructorName(m.Constructor)))
	// 	switch m.Constructor {
	// 	case rony.C_Error:
	// 		success = false
	// 		x := &rony.Error{}
	// 		if err := x.Unmarshal(m.Message); err != nil {
	// 			logs.Error("We couldn't unmarshal MessagesSendMedia (Error) response", zap.Error(err))
	// 		}
	// 		logs.Error("We received error on MessagesSendMedia response",
	// 			zap.String("Code", x.Code),
	// 			zap.String("Item", x.Items),
	// 		)
	// 		if x.Code == msg.ErrCodeAlreadyExists && x.Items == msg.ErrItemRandomID {
	// 			success = true
	// 			_ = repo.PendingMessages.Delete(uploadRequest.MessageID)
	//
	// 		}
	// 	}
	// 	waitGroup.Done()
	// }
	// timeoutCB := func() {
	// 	success = false
	// 	logs.Debug("We got Timeout! on MessagesSendMedia response")
	// 	waitGroup.Done()
	// }
	// r.queueCtrl.EnqueueCommand(
	// 	&rony.MessageEnvelope{
	// 		Constructor: msg.C_MessagesSendMedia,
	// 		RequestID:   uint64(x.RandomID),
	// 		Message:     reqBuff,
	// 		Header:      domain.TeamHeader(pendingMessage.TeamID, pendingMessage.TeamAccessHash),
	// 	},
	// 	timeoutCB, successCB, false)
	waitGroup.Wait()
	return
}

func (r *River) registerCommandHandlers() {
	r.localCommands = map[int64]domain.LocalMessageHandler{
		msg.C_AccountGetTeams:        r.accountsGetTeams,
		msg.C_ClientContactSearch:    r.clientContactSearch,
		msg.C_ClientGetRecentSearch:  r.clientGetRecentSearch,
		msg.C_ClientGetTeamCounters:  r.clientGetTeamCounters,
		msg.C_ClientGlobalSearch:     r.clientGlobalSearch,
		msg.C_ClientPutRecentSearch:  r.clientPutRecentSearch,
		msg.C_ClientSendMessageMedia: r.clientSendMessageMedia,
		msg.C_MessagesDelete:         r.messagesDelete,
		msg.C_MessagesGet:            r.messagesGet,
		msg.C_MessagesGetDialog:      r.messagesGetDialog,
		msg.C_MessagesGetDialogs:     r.messagesGetDialogs,
		msg.C_MessagesSendMedia:      r.messagesSendMedia,
		msg.C_UsersGet:               r.usersGet,
		msg.C_UsersGetFull:           r.usersGetFull,
	}
}

// RiverConnection connection info
type RiverConnection struct {
	AuthID  int64
	AuthKey [256]byte
	UserID  int64
}
