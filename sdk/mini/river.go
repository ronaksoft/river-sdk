package mini

import (
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	networkCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_network"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"github.com/ronaksoft/rony"
	"go.uber.org/zap"
	"strings"
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

type LocalHandler func(in, out *rony.MessageEnvelope, da *DelegateAdapter)

type RiverConfig struct {
	ServerHostPort string
	// DbPath is the path of the folder holding the sqlite database.
	DbPath string
	// DbID is used to save data for different accounts in separate databases. Could be used for multi-account cases.
	DbID string
	// MainDelegate holds all the general callback functions that let the user of this SDK
	// get notified of the events.
	MainDelegate MainDelegate

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
	localCommands map[int64]LocalHandler

	// Internal Controllers
	networkCtrl *networkCtrl.Controller

	// Delegates
	mainDelegate MainDelegate
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
	r.mainDelegate = conf.MainDelegate

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
	repo.SetRootFolders(
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

func (r *River) registerCommandHandlers() {
	r.localCommands = map[int64]LocalHandler{
		msg.C_ClientSendMessageMedia: r.clientSendMessageMedia,
		msg.C_ClientGlobalSearch:     r.clientGlobalSearch,
		msg.C_MessagesSendMedia:      r.messagesSendMedia,
		msg.C_MessagesGetDialogs:     r.messagesGetDialogs,
	}
}

// RiverConnection connection info
type RiverConnection struct {
	AuthID  int64
	AuthKey []byte
	UserID  int64
}
