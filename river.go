package riversdk

import (
	"sync"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/network"
	"git.ronaksoftware.com/ronak/riversdk/queue"
	"git.ronaksoftware.com/ronak/riversdk/synchronizer"
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
	delegateMutex sync.Mutex
	delegates     map[int64]RequestDelegate
	localCommands map[int64]domain.LocalMessageHandler

	// Internal Controllers
	networkCtrl *network.NetworkController
	queueCtrl   *queue.QueueController
	syncCtrl    *synchronizer.SyncController

	// RealTimeRequests is list of requests that should not passed to queue to send they should directly pass to networkController
	realTimeRequest map[int64]bool

	mainDelegate MainDelegate

	// Connection Info
	ConnInfo *RiverConnection
}
