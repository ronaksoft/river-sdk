package riversdk

import (
	"context"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_network"
	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_queue"
	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_sync"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
)

// RiverConfig
type RiverConfig struct {
	ServerEndpoint string
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
	// ServerKeysFilePath is the path of a json file holding finger print and public keys.
	ServerKeysFilePath string
	// MainDelegate holds all the general callback functions that let the user of this SDK
	// get notified of the events.
	MainDelegate MainDelegate

	// LogLevel
	LogLevel int

	// Logger pass logs to external handler
	Logger LoggerDelegate

	// Folder path to save files
	DocumentPhotoDirectory string
	DocumentVideoDirectory string
	DocumentFileDirectory  string
	DocumentAudioDirectory string
	DocumentCacheDirectory string
	DocumentLogDirectory   string
	// Connection Info
	ConnInfo *RiverConnection
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
	networkCtrl *network.Controller
	queueCtrl   *queue.Controller
	syncCtrl    *synchronizer.Controller

	// Delegates
	delegateMutex sync.Mutex
	delegates     map[int64]RequestDelegate
	mainDelegate  MainDelegate
	logger        LoggerDelegate

	// implements wait 500 ms on out of sync to receive possible missed updates
	lastOutOfSyncTime  time.Time
	chOutOfSyncUpdates chan []*msg.UpdateContainer
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
	networkCtrl := network.NewController(
		network.Config{
			ServerEndpoint: url,
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
		_ = networkCtrl.Send(msgEnvelope, true)
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
	go networkCtrl.Connect()
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
