package riversdk

import (
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"github.com/dustin/go-humanize"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"testing"
)

var (
	_River *River
)

type MainDelegateDummy struct{}

func (d *MainDelegateDummy) OnUpdates(constructor int64, b []byte) {}

func (d *MainDelegateDummy) OnDeferredRequests(requestID int64, b []byte) {}

func (d *MainDelegateDummy) OnNetworkStatusChanged(quality int) {
	state := domain.NetworkStatus(quality)
	logs.Info("Network status changed", zap.String("Status", domain.NetworkStatusName[state]))
}

func (d *MainDelegateDummy) OnSyncStatusChanged(newStatus int) {
	state := domain.SyncStatus(newStatus)
	logs.Info("Sync status changed", zap.String("Status", domain.SyncStatusName[state]))
}

func (d *MainDelegateDummy) OnAuthKeyCreated(authID int64) {
	logs.Info("Auth Key Created", zap.Int64("AuthID", authID))
}

func (d *MainDelegateDummy) OnGeneralError(b []byte) {
	e := new(msg.Error)
	e.Unmarshal(b)
	logs.Error("Received general error", zap.String("Code", e.Code), zap.String("Items", e.Items))
}

func (d *MainDelegateDummy) OnSessionClosed(res int) {
	logs.Message("Session Closed", zap.Int("Res", res))
}

func (d *MainDelegateDummy) OnDownloadProgressChanged(messageID, processedParts, totalParts int64, percent float64) {
	logs.Message("Download progress changed", zap.Float64("Progress", percent))
}

func (d *MainDelegateDummy) OnUploadProgressChanged(messageID, processedParts, totalParts int64, percent float64) {
	logs.Message("Upload progress changed", zap.Float64("Progress", percent))
}

func (d *MainDelegateDummy) OnDownloadCompleted(messageID int64, filePath string) {
	logs.Info("Download completed", zap.Int64("MsgID", messageID), zap.String("FilePath", filePath))
}

func (d *MainDelegateDummy) OnUploadCompleted(messageID int64, filePath string) {
	logs.Info("Upload completed", zap.Int64("MsgID", messageID), zap.String("FilePath", filePath))
}

func (d *MainDelegateDummy) OnUploadError(messageID, requestID int64, filePath string, err []byte) {
	x := new(msg.Error)
	x.Unmarshal(err)

	logs.Error("OnUploadError",
		zap.String("Code", x.Code),
		zap.String("Item", x.Items),
		zap.Int64("MsgID", messageID),
		zap.Int64("ReqID", requestID),
		zap.String("FilePath", filePath),
	)

}

func (d *MainDelegateDummy) OnDownloadError(messageID, requestID int64, filePath string, err []byte) {
	x := new(msg.Error)
	x.Unmarshal(err)

	logs.Error("OnDownloadError",
		zap.String("Code", x.Code),
		zap.String("Item", x.Items),
		zap.Int64("MsgID", messageID),
		zap.Int64("ReqID", requestID),
		zap.String("FilePath", filePath),
	)
}


func init() {
	logs.Info("Creating New River SDK Instance")
	r := new(River)
	r.SetConfig(&RiverConfig{
		DbPath:             "./_data/",
		DbID:               "test",
		ServerKeysFilePath: "./keys.json",
		ServerEndpoint:     "ws://new.river.im",
		QueuePath:              fmt.Sprintf("%s/%s", "./_queue", "test"),
		MainDelegate:           new(MainDelegateDummy),
		Logger:                 nil,
		LogLevel:               int(zapcore.DebugLevel),
		DocumentAudioDirectory: "./_files/audio",
		DocumentVideoDirectory: "./_files/video",
		DocumentPhotoDirectory: "./_files/photo",
		DocumentFileDirectory:  "./_files/file",
		DocumentCacheDirectory: "./_files/cache",
		DocumentLogDirectory:   "./_files/logs",
	})

	// r.Start()
	// if r.ConnInfo.AuthID == 0 {
	// 	logs.Info("AuthKey has not been created yet.")
	// 	if err := r.CreateAuthKey(); err != nil {
	// 		return
	// 	}
	// 	logs.Info("AuthKey Created.")
	// }
	_River = r
}

func dummyMessageHandler(m *msg.MessageEnvelope) {
	logs.Info("MessageEnvelope Handler",
		zap.Uint64("REQ_ID", m.RequestID),
		zap.String("CONSTRUCTOR", msg.ConstructorNames[m.Constructor]),
		zap.String("LENGTH", humanize.Bytes(uint64(len(m.Message)))),
	)
}

func dummyUpdateHandler(u *msg.UpdateContainer) {
	logs.Info("UpdateContainer Handler",
		zap.Int64("MIN_UPDATE_ID", u.MinUpdateID),
		zap.Int64("MAX_UPDATE_ID", u.MaxUpdateID),
		zap.Int32("LENGTH", u.Length),
	)
	for _, update := range u.Updates {
		logs.Info("UpdateEnvelope",
			zap.String("CONSTRUCTOR", msg.ConstructorNames[update.Constructor]),
		)
	}
}

// func dummyQualityUpdateHandler(q NetworkStatus) {
// 	logs.Info("Network Quality Updated",
// 		zap.Int("New Quality", int(q)),
// 	)
// }
//
// func newTestNetwork(pintTime time.Duration) *networkController {
// 	ctlConfig := networkConfig{
// 		ServerEndpoint: "ws://river.im",
// 		PingTime:       pintTime,
// 	}
//
// 	ctl := newNetworkController(ctlConfig)
// 	return ctl
// }
//
// func newTestQueue(network *networkController) *queueController {
// 	q, err := newQueueController(network, "./_data/queue", nil)
// 	if err != nil {
// 		logs.Fatal(err.Error())
// 	}
// 	return q
// }
//
// func newTestDB() *sqlx.DB {
// 	// Initialize Database directory (LevelDB)
// 	dbPath := "./_data/"
// 	dbID := "test"
// 	os.MkdirAll(dbPath, os.ModePerm)
// 	if db, err := sqlx.Open("sqlite3", fmt.Sprintf("%s/%s.db", dbPath, dbID)); err != nil {
// 		logs.Fatal(err.Error())
// 	} else {
// 		setDB(db)
// 		return db
// 	}
// 	return nil
// }
//
// func newTestSync(queue *queueController) *syncController {
// 	s := newSyncController(
// 		syncConfig{
// 			QueueCtrl: queue,
// 		},
// 	)
// 	return s
// }
//
// func TestNewSyncController(t *testing.T) {
// 	setDB(newTestDB())
// 	networkCtrl := newTestNetwork(20 * time.Second)
// 	queueCtrl := newTestQueue(networkCtrl)
// 	syncCtrl := newTestSync(queueCtrl)
//
// 	networkCtrl.SetMessageHandler(queueCtrl.messageHandler)
// 	networkCtrl.SetUpdateHandler(syncCtrl.updateHandler)
// 	networkCtrl.SetNetworkStatusChangedCallback(dummyQualityUpdateHandler)
//
// 	networkCtrl.Start()
// 	queueCtrl.Start()
// 	syncCtrl.Start()
// 	networkCtrl.Connect()
//
// 	for i := 0; i < 100; i++ {
// 		requestID := _RandomUint64()
// 		initConn := new(msg.InitConnect)
// 		initConn.ClientNonce = _RandomUint64()
// 		messageEnvelope := new(msg.MessageEnvelope)
// 		messageEnvelope.Constructor = msg.C_InitConnect
// 		messageEnvelope.RequestID = requestID
// 		messageEnvelope.Message, _ = initConn.Marshal()
// 		req := request{
// 			Timeout:         2 * time.Second,
// 			MessageEnvelope: messageEnvelope,
// 		}
// 		logs.Info("prepare request",
// 			zap.Uint64("ReqID", messageEnvelope.RequestID),
// 		)
// 		queueCtrl.addToWaitingList(&req)
// 	}
//
// 	time.Sleep(30 * time.Second)
//
// }
//
// func TestNewNetworkController(t *testing.T) {
// 	ctl := newTestNetwork(time.Second)
// 	ctl.SetMessageHandler(dummyMessageHandler)
// 	ctl.SetUpdateHandler(dummyUpdateHandler)
// 	ctl.SetNetworkStatusChangedCallback(dummyQualityUpdateHandler)
// 	if err := ctl.Start(); err != nil {
// 		logs.Info(err.Error())
// 		return
// 	}
// 	ctl.Connect()
//
// 	bufio.NewReader(os.Stdin).ReadBytes('\n')
// 	ctl.Disconnect()
// 	ctl.Stop()
// }

func TestNewRiver(t *testing.T) {
	logs.Info("Creating New River SDK Instance")
	r := new(River)
	r.SetConfig(&RiverConfig{
		DbPath:             "./_data/",
		DbID:               "test",
		ServerKeysFilePath: "./keys.json",
		ServerEndpoint:     "ws://river.im",
	})

	r.Start()
	if r.ConnInfo.AuthID == 0 {
		logs.Info("AuthKey has not been created yet.")
		if err := r.CreateAuthKey(); err != nil {
			t.Error(err.Error())
			return
		}
		logs.Info("AuthKey Created.")
	}
}

func TestRiver_SetScrollStatus(t *testing.T) {
	peerID := int64(100)
	peerType := int32(2)
	msgID := int64(101)
	_River.SetScrollStatus(peerID, msgID, peerType)

	readMsgID := _River.GetScrollStatus(peerID, peerType)
	if readMsgID != msgID {
		t.Error("values do not match")
	}
}
