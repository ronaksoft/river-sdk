package riversdk

import (
	"encoding/json"
	"fmt"
	"git.ronaksoftware.com/river/msg/chat"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"os"
	"testing"
	"time"
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
		// _River.syncCtrl.GetServerSalt()
		time.Sleep(time.Second * 5)
	})
}

func init() {
	logs.Info("Creating New River SDK Instance")
	r := new(River)
	conInfo := new(RiverConnection)
	connDelegate := &ConnInfoDelegates{}
	connDelegate.Load(conInfo)
	conInfo.Delegate = new(ConnInfoDelegates)

	serverKeyBytes, _ := ioutil.ReadFile("./keys.json")
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
	e := new(msg.Error)
	e.Unmarshal(b)
	logs.Error("Received general error", zap.String("Code", e.Code), zap.String("Items", e.Items))
}

func (d *MainDelegateDummy) OnSessionClosed(res int) {
	logs.Info("Session Closed", zap.Int("Res", res))
}

func (d *MainDelegateDummy) ShowLoggerAlert() {}

func (d *MainDelegateDummy) AddLog(txt string) {}

func (d *MainDelegateDummy) AppUpdate(version string, available, force bool) {}

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
