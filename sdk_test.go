package riversdk

import (
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"os"
	"testing"
)

var (
	_River *River
)

func TestSDK(t *testing.T) {
	Convey("Start", t, func(c C) {
		err := _River.Start()
		c.So(err, ShouldBeNil)
		_River.StartNetwork()
	})
	Convey("Check Reconnect", t, func() {})
}

func init() {
	logs.Info("Creating New River SDK Instance")
	r := new(River)
	conInfo := new(RiverConnection)
	conInfo.Delegate = new(dummyConInfoDelegate)
	serverKeyBytes, _ := ioutil.ReadFile("./keys.json")
	r.SetConfig(&RiverConfig{
		DbPath:                 "./_data/",
		DbID:                   "test",
		QueuePath:              fmt.Sprintf("%s/%s", "./_queue", "test"),
		MainDelegate:           new(MainDelegateDummy),
		FileDelegate:           new(FileDelegateDummy),
		LogLevel:               int(zapcore.DebugLevel),
		DocumentAudioDirectory: "./_files/audio",
		DocumentVideoDirectory: "./_files/video",
		DocumentPhotoDirectory: "./_files/photo",
		DocumentFileDirectory:  "./_files/file",
		DocumentCacheDirectory: "./_files/cache",
		DocumentLogDirectory:   "./_files/logs",
		ConnInfo:               conInfo,
		ServerKeys:             string(serverKeyBytes),
		ServerEndpoint:         "ws://river.ronaksoftware.com",
		FileServerEndpoint:     "http://river.ronaksoftware.com:8080",
	})
	_River = r

	repo.InitRepo(fmt.Sprintf("%s/%s.db", "./_data", "test"), false)
}

type ConnInfoDelegates struct{}

func (c *ConnInfoDelegates) SaveConnInfo(connInfo []byte) {
	_ = os.MkdirAll("./_connection", os.ModePerm)
	err := ioutil.WriteFile("./_connection/connInfo", connInfo, 0666)
	if err != nil {
		fmt.Println(err)
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

func (d *MainDelegateDummy) ShowLoggerAlert() {}

func (d *MainDelegateDummy) AddLog(txt string) {}

type RequestDelegateDummy struct{}

func (RequestDelegateDummy) OnComplete(b []byte) {
	fmt.Println(b)
}

func (RequestDelegateDummy) OnTimeout(err error) {
	fmt.Println(err)
}

type FileDelegateDummy struct{}

func (d *FileDelegateDummy) OnProgressChanged(reqID string, clusterID int32, fileID, accessHash, percent int64) {
	logs.Info("Download progress changed", zap.Int64("Progress", percent))
}

func (d *FileDelegateDummy) OnCompleted(reqID string, clusterID int32, fileID, accessHash int64, filePath string) {
	logs.Info("Download completed", zap.String("ReqID", reqID), zap.String("FilePath", filePath))

}

func (d *FileDelegateDummy) OnCancel(reqID string, clusterID int32, fileID, accessHash int64, hasError bool) {
	logs.Error("OnCancel")
}
