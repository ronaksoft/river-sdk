package riversdk

import (
	"encoding/json"
	"fmt"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"github.com/ronaksoft/rony"
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
		err := _River.AppStart()
		c.So(err, ShouldBeNil)

		_River.StartNetwork("")
		if _River.ConnInfo.AuthID == 0 {
			_River.CreateAuthKey()
		}
	})
	Convey("Get Salt", t, func() {

	})
}

func init() {
	logs.Info("Creating New River SDK Instance")
	r := new(River)
	conInfo := new(RiverConnection)
	connDelegate := &ConnInfoDelegates{}
	connDelegate.Load(conInfo)
	conInfo.Delegate = new(ConnInfoDelegates)

	r.SetConfig(&RiverConfig{
		DbPath:                 "./_data/",
		DbID:                   "test",
		MainDelegate:           new(MainDelegateDummy),
		FileDelegate:           new(FileDelegateDummy),
		LogLevel:               int(zapcore.DebugLevel),
		DocumentAudioDirectory: "./_files/audio",
		DocumentVideoDirectory: "./_files/video",
		DocumentPhotoDirectory: "./_files/photo",
		DocumentFileDirectory:  "./_files/file",
		DocumentCacheDirectory: "./_files/cache",
		LogDirectory:           "./_files/logs",
		ConnInfo:               conInfo,
		SeedHostPorts:          []string{"river.ronaksoftware.com"},
	})
	_River = r

	repo.MustInit(fmt.Sprintf("%s/%s.db", "./_data", "test"), false)
}

type ConnInfoDelegates struct{}

func (c *ConnInfoDelegates) Load(connInfo *RiverConnection) {
	b, _ := ioutil.ReadFile("./_connection/connInfo")
	_ = json.Unmarshal(b, connInfo)
	return
}

func (c *ConnInfoDelegates) SaveConnInfo(connInfo []byte) {
	_ = os.MkdirAll("./_connection", os.ModePerm)
	err := ioutil.WriteFile("./_connection/connInfo", connInfo, 0666)
	if err != nil {
		fmt.Println(err)
	}
}

func (c *ConnInfoDelegates) Get(key string) string { return "" }

func (c *ConnInfoDelegates) Set(key, value string) {}

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
	e := new(rony.Error)
	e.Unmarshal(b)
	logs.Error("Received general error", zap.String("Code", e.Code), zap.String("Items", e.Items))
}

func (d *MainDelegateDummy) OnSessionClosed(res int) {
	logs.Info("Session Closed", zap.Int("Res", res))
}

func (d *MainDelegateDummy) ShowLoggerAlert() {}

func (d *MainDelegateDummy) AddLog(txt string) {}

func (d *MainDelegateDummy) AppUpdate(version string, available, force bool) {}

func (d *MainDelegateDummy) DataSynced(dialogs, contacts, gifs bool) {
}

type RequestDelegateDummy struct{}

func (RequestDelegateDummy) OnComplete(b []byte) {
	fmt.Println(b)
}

func (RequestDelegateDummy) OnTimeout(err error) {
	fmt.Println(err)
}

type FileDelegateDummy struct{}

func (d *FileDelegateDummy) OnProgressChanged(reqID string, clusterID int32, fileID, accessHash, percent int64, peerID int64) {
	logs.Info("Download progress changed",
		zap.Int64("Progress", percent),
		zap.Int64("PeerID", peerID),
	)
}

func (d *FileDelegateDummy) OnCompleted(reqID string, clusterID int32, fileID, accessHash int64, filePath string, peerID int64) {
	logs.Info("Download completed",
		zap.String("ReqID", reqID),
		zap.String("FilePath", filePath),
		zap.Int64("PeerID", peerID),
	)

}

func (d *FileDelegateDummy) OnCancel(reqID string, clusterID int32, fileID, accessHash int64, hasError bool, peerID int64) {
	logs.Error("CancelCB")
}
