package riversdk

import (
	"context"
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	fileCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_file"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"strings"
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_network"
	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_queue"
	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_sync"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
)

// RiverConfig
type RiverConfig struct {
	ServerEndpoint     string
	FileServerEndpoint string
	// PingTimeSec sets how often a ping message will be sent to the server. Ping messages
	// are used to calculate the quality of the network.
	PingTimeSec int32
	// PongTimeoutSec is the amount of time in seconds which SDK will wait after sending
	// a ping to server to get the pong back. If it does not receive the pong message in
	// this period of time, it disconnects and reconnect.
	PongTimeoutSec int32
	// QueuePath is the path of a folder that pending requests will be saved there until sending
	// to the server.
	QueuePath string
	// DbPath is the path of the folder holding the sqlite database.
	DbPath string
	// DbID is used to save data for different accounts in separate databases.
	DbID string
	// ServerKeysFilePath is a json file holding finger print and public keys.
	ServerKeys string
	// MainDelegate holds all the general callback functions that let the user of this SDK
	// get notified of the events.
	MainDelegate MainDelegate
	// FileDelegate holds all the callbacks required by file related functions
	FileDelegate FileDelegate

	// LogLevel
	LogLevel int

	// Folder path to save files
	DocumentPhotoDirectory string
	DocumentVideoDirectory string
	DocumentFileDirectory  string
	DocumentAudioDirectory string
	DocumentCacheDirectory string
	DocumentLogDirectory   string
	// Connection Info
	ConnInfo *RiverConnection

	// ClientInfo
	ClientPlatform string
	ClientVersion  string

	// OptimizeForLowMemory if is set then SDK tries to use the lowest possible ram
	OptimizeForLowMemory bool
}

// River
// This the main and a wrapper around all the components of the system (networkController, queueController,
// syncController). All the controllers could be used standalone, but this SDK connect them in a way
// we think is the best possible.
// Only the functions which are exposed will be used by the user of the SDK. All the low-level tasks
// to smooth the connection between client and server are done by this SDK. The underlying storage used
// by this SDK is Sqlite3, however with the least effort you can use other SQL databases. 'model' is the
// package name selected to handle repository functions.
type River struct {
	// Connection Info
	ConnInfo *RiverConnection
	// Device Token
	DeviceToken *msg.AccountRegisterDevice

	// localCommands can be satisfied by client cache
	localCommands map[int64]domain.LocalMessageHandler
	// realTimeCommands should not passed to queue to send they should directly pass to networkController
	realTimeCommands map[int64]bool

	// Internal Controllers
	networkCtrl *networkCtrl.Controller
	queueCtrl   *queueCtrl.Controller
	syncCtrl    *syncCtrl.Controller
	fileCtrl    *fileCtrl.Controller

	// Delegates
	delegateMutex sync.Mutex
	delegates     map[int64]RequestDelegate
	mainDelegate  MainDelegate
	fileDelegate  FileDelegate

	// implements wait 500 ms on out of sync to receive possible missed updates
	lastOutOfSyncTime    time.Time
	chOutOfSyncUpdates   chan []*msg.UpdateContainer
	dbPath               string
	optimizeForLowMemory bool
}

// SetConfig ...
// This function must be called before any other function, otherwise it panics
func (r *River) SetConfig(conf *RiverConfig) {
	viper.Set(ConfClientPlatform, conf.ClientPlatform)
	viper.Set(ConfClientVersion, conf.ClientVersion)

	r.lastOutOfSyncTime = time.Now().Add(1 * time.Second)
	r.chOutOfSyncUpdates = make(chan []*msg.UpdateContainer, 500)
	r.ConnInfo = conf.ConnInfo
	r.optimizeForLowMemory = conf.OptimizeForLowMemory

	// Initialize DB Path
	if strings.HasPrefix(conf.DbPath, "file://") {
		conf.DbPath = conf.DbPath[7:]
	}
	conf.DbPath = strings.TrimRight(conf.DbPath, "/ ")
	r.dbPath = fmt.Sprintf("%s/%s.db", conf.DbPath, conf.DbID)


	r.registerCommandHandlers()
	r.delegates = make(map[int64]RequestDelegate)
	r.mainDelegate = conf.MainDelegate
	r.fileDelegate = conf.FileDelegate

	// set loglevel
	logs.SetLogLevel(conf.LogLevel)

	// set log file path
	if conf.DocumentLogDirectory != "" {
		_ = logs.SetLogFilePath(conf.DocumentLogDirectory)
	}

	// Initialize realtime requests
	r.realTimeCommands = map[int64]bool{
		msg.C_MessagesSetTyping: true,
	}

	// Initialize Network Controller
	r.networkCtrl = networkCtrl.New(
		networkCtrl.Config{
			WebsocketEndpoint: conf.ServerEndpoint,
			HttpEndpoint:      conf.FileServerEndpoint,
			PingTime:          time.Duration(conf.PingTimeSec) * time.Second,
			PongTimeout:       time.Duration(conf.PongTimeoutSec) * time.Second,
		},
	)
	r.networkCtrl.SetNetworkStatusChangedCallback(func(newQuality domain.NetworkStatus) {
		if r.mainDelegate != nil {
			r.mainDelegate.OnNetworkStatusChanged(int(newQuality))
		}
	})
	r.networkCtrl.SetErrorHandler(r.onGeneralError)
	r.networkCtrl.SetMessageHandler(r.onReceivedMessage)
	r.networkCtrl.SetUpdateHandler(r.onReceivedUpdate)
	r.networkCtrl.SetOnConnectCallback(r.onNetworkConnect)
	r.networkCtrl.SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey[:])

	// Initialize FileController
	fileCtrl.SetRootFolders(conf.DocumentAudioDirectory, conf.DocumentFileDirectory, conf.DocumentPhotoDirectory, conf.DocumentVideoDirectory, conf.DocumentCacheDirectory)
	r.fileCtrl = fileCtrl.New(r.networkCtrl,
		fileCtrl.Config{
			OnUploadCompleted:   r.onFileUploadCompleted,
			ProgressCallback:    r.onFileProgressChanged,
			OnDownloadCompleted: r.onFileDownloadCompleted,
			OnFileUploadError:   r.onFileUploadError,
			OnFileDownloadError: r.onFileDownloadError,
		},
	)

	// Initialize queueController
	if q, err := queueCtrl.New(r.networkCtrl, conf.QueuePath); err != nil {
		logs.Fatal("River::SetConfig() faild to initialize MessageQueue",
			zap.String("Error", err.Error()),
		)
	} else {
		r.queueCtrl = q
	}

	// Initialize Sync Controller
	r.syncCtrl = syncCtrl.NewSyncController(
		syncCtrl.Config{
			ConnInfo:    r.ConnInfo,
			NetworkCtrl: r.networkCtrl,
			QueueCtrl:   r.queueCtrl,
			FileCtrl:    r.fileCtrl,
		},
	)

	// call external delegate on sync status changed
	r.syncCtrl.SetSyncStatusChangedCallback(func(newStatus domain.SyncStatus) {
		if r.mainDelegate != nil {
			r.mainDelegate.OnSyncStatusChanged(int(newStatus))
		}
	})
	// call external delegate on OnUpdate
	r.syncCtrl.SetOnUpdateCallback(func(constructorID int64, b []byte) {
		if r.mainDelegate != nil {
			r.mainDelegate.OnUpdates(constructorID, b)
		}
	})

	// Initialize Server Keys
	if err := _ServerKeys.UnmarshalJSON([]byte(conf.ServerKeys)); err != nil {
		logs.Fatal("River::SetConfig() failed to unmarshal server keys",
			zap.String("Error", err.Error()),
		)
	}

	// Initialize River Connection
	logs.Info("River::SetConfig() Load/Create New River Connection")

	if r.ConnInfo.UserID != 0 {
		r.syncCtrl.SetUserID(r.ConnInfo.UserID)
	}
}

func (r *River) Version() string {
	// TODO:: automatic generation
	return "0.8.1"
}

// GetWorkGroup
// Client call GetWorkGroup with a timeout set, if this function could connect to server and get its response back from
// the server then it returns the serialized version of msg.SystemInfo, otherwise it returns an error
// It is upto the caller to re-call this function in case of error returned.
func GetWorkGroup(url string, timeoutSecond int) ([]byte, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(timeoutSecond)*time.Second)
	defer cancelFunc()

	b, err := getWorkGroup(ctx, url)
	return b, err
}

func getWorkGroup(ctx context.Context, url string) ([]byte, error) {
	networkCtrl := networkCtrl.New(
		networkCtrl.Config{
			WebsocketEndpoint: url,
		},
	)

	ch := make(chan []byte)
	// Assign Handlers
	// OnConnect Handler
	networkCtrl.SetOnConnectCallback(func() {
		msgEnvelope := new(msg.MessageEnvelope)
		msgEnvelope.RequestID = ronak.RandomUint64()
		msgEnvelope.Constructor = msg.C_SystemGetInfo

		msgEnvelope.Message, _ = (&msg.SystemGetInfo{}).Marshal()
		_ = networkCtrl.SendWebsocket(msgEnvelope, true)
	})
	// Message Handler
	networkCtrl.SetMessageHandler(func(messages []*msg.MessageEnvelope) {
		for _, message := range messages {
			switch message.Constructor {
			case msg.C_SystemInfo:
				ch <- message.Message
				return
			}
		}
	})
	// Update Handler
	networkCtrl.SetUpdateHandler(func(messages []*msg.UpdateContainer) {
		// We don't need to handle updates
		return
	})

	// Start the Network Controller alone
	_ = networkCtrl.Start()
	go networkCtrl.Connect(false)
	defer networkCtrl.Stop()       // 2nd Stop the controller
	defer networkCtrl.Disconnect() // 1st Disconnect

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case b := <-ch:
			return b, nil
		}
	}

}
