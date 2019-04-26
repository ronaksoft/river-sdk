package riversdk

import (
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"testing"
)

var (
	_River *River
)

func init() {
	logs.Info("Creating New River SDK Instance")
	r := new(River)
	r.SetConfig(&RiverConfig{
		DbPath:                 "./_data/",
		DbID:                   "test",
		ServerKeysFilePath:     "./keys.json",
		ServerEndpoint:         "ws://new.river.im",
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

	r.Start()
	if r.ConnInfo.AuthID == 0 {
		logs.Info("AuthKey has not been created yet.")
		if err := r.CreateAuthKey(); err != nil {
			return
		}
		logs.Info("AuthKey Created.")
	}
	_River = r
}

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

type MainDelegateDummy struct{}

func (d *MainDelegateDummy) OnUpdates(constructor int64, b []byte) {}

func (d *MainDelegateDummy) OnDeferredRequests(requestID int64, b []byte) {}

func (d *MainDelegateDummy) OnNetworkStatusChanged(quality int) {
	state := domain.NetworkStatus(quality)
	logs.Info("Network status changed", zap.String("Status", state.ToString()))
}

func (d *MainDelegateDummy) OnSyncStatusChanged(newStatus int) {
	state := domain.SyncStatus(newStatus)
	logs.Info("Sync status changed", zap.String("Status", state.ToString()))
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
	logs.Info("Session Closed", zap.Int("Res", res))
}

func (d *MainDelegateDummy) OnDownloadProgressChanged(messageID, processedParts, totalParts int64, percent float64) {
	logs.Info("Download progress changed", zap.Float64("Progress", percent))
}

func (d *MainDelegateDummy) OnUploadProgressChanged(messageID, processedParts, totalParts int64, percent float64) {
	logs.Info("Upload progress changed", zap.Float64("Progress", percent))
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
