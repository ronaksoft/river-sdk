package riversdk

import (
	"encoding/json"
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"git.ronaksoftware.com/ronak/riversdk/pkg/salt"
	"git.ronaksoftware.com/ronak/toolbox"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	_River *River
)

func TestSDK(t *testing.T) {
	Convey("Check Salt", t, func() {
		repo.InitRepo(fmt.Sprintf("%s/%s.db", "./_data", "test"), false)
		var saltArrays [][]domain.Slt
		var saltArray []domain.Slt
		ti := time.Now()
		for i := 0; i < 48; i++ {
			slt := domain.Slt{}
			next := ti.Add(time.Hour * time.Duration(i))
			slt.Timestamp = time.Unix(next.Unix(), 0).Unix()
			slt.Value = ronak.RandomInt64(0)
			if i == 0 {
				slt.Value = 5555
			}
			saltArray = append(saltArray, slt)
		}
		saltArrays = append(saltArrays, saltArray)

		var saltArray2 []domain.Slt
		for i := 0; i < 48; i++ {
			slt := domain.Slt{}
			next := ti.Add(time.Hour * time.Duration(i*48))
			slt.Timestamp = time.Unix(next.Unix(), 0).Unix()
			slt.Value = ronak.RandomInt64(0)
			saltArray2 = append(saltArray2, slt)
		}
		saltArrays = append(saltArrays, saltArray2)
		tests := []struct {
			name  string
			salts []domain.Slt
		}{
			{"test1", saltArrays[0]},
			{"test2", saltArrays[1]},
		}
		for i, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				b, _ := json.Marshal(tt.salts)
				err := repo.System.SaveString(domain.SkSystemSalts, string(b))
				if err != nil {
					logs.Debug("synchronizer::SystemGetSalts()",
						zap.String("error", err.Error()),
					)
				}
				time.Sleep(time.Millisecond * 600)

				salt.UpdateSalt()
				s := _River.GetSDKSalt()
				if i == 0 {
					if s != 5555 {
						t.Error(fmt.Sprintf("expecting 5555, have %d", s))
					}
				}
				if i == 1 {
					logs.Debug("s::()",
						zap.Int64("s", s),
					)
				}
			})
		}
	})
	Convey("Check GetWorkGroup", t, func() {
		b, err := GetWorkGroup("ws://alaki.river.im", 5)
		So(err, ShouldBeNil)

		si := new(msg.SystemInfo)
		err = si.Unmarshal(b)
		So(err, ShouldBeNil)
		t.Log("WorkGroupName:", si.WorkGroupName)
	})
	Convey("Check Reconnect", t, func() {
		logs.SetLogLevel(0)
		fmt.Println("Creating New River SDK Instance")
		conInfo := new(RiverConnection)

		file, err := os.Open("./_connection/connInfo1")
		if err == nil {
			b, _ := ioutil.ReadAll(file)
			err := json.Unmarshal(b, conInfo)
			So(err, ShouldBeNil)
		}

		conInfo.Delegate = new(ConnInfoDelegates)

		r := new(River)
		fmt.Println("SetConfig called")
		r.SetConfig(&RiverConfig{
			DbPath:    "./_data/",
			DbID:      "test",
			QueuePath: "./_queue/",
			// ServerKeysFilePath: "./keys.json",

			ServerEndpoint: "ws://new.river.im",
			LogLevel:       0,
			ConnInfo:       conInfo,
		})

		fmt.Println("Start called")
		_ = r.Start()
		for r.ConnInfo.AuthID == 0 {
			err := r.CreateAuthKey()
			So(err, ShouldBeNil)
		}

		time.Sleep(10 * time.Second)
		r.ResetAuthKey()
		r.stop()

		time.Sleep(10 * time.Second)

		// Connect to 2nd Server
		file, err = os.Open("./_connection/connInfo2")
		if err == nil {
			b, _ := ioutil.ReadAll(file)
			err := json.Unmarshal(b, conInfo)
			So(err, ShouldBeNil)
		}

		conInfo.Delegate = new(ConnInfoDelegates)

		r.SetConfig(&RiverConfig{
			DbPath:    "./_data/",
			DbID:      "test",
			QueuePath: "./_queue/",
			// ServerKeysFilePath: "./keys.json",
			ServerEndpoint: "ws://new.river.im",
			ConnInfo:       conInfo,
			LogLevel:       0,
		})
		_ = r.Start()
		for r.ConnInfo.AuthID == 0 {
			logs.Info("AuthKey has not been created yet.")
			err := r.CreateAuthKey()
			So(err, ShouldBeNil)
			logs.Info("AuthKey Created.")
		}
		fmt.Println("AuthID", r.ConnInfo.AuthID)
		fmt.Println("AuthKey", r.ConnInfo.AuthKey)
		time.Sleep(time.Second * 10)
		b := r.GetMonitorStats()
		fmt.Println(string(b))
	})
}

func init() {
	logs.Info("Creating New River SDK Instance")
	r := new(River)
	conInfo := new(RiverConnection)
	conInfo.Delegate = new(dummyConInfoDelegate)
	r.SetConfig(&RiverConfig{
		DbPath:                 "./_data/",
		DbID:                   "test",
		QueuePath:              fmt.Sprintf("%s/%s", "./_queue", "test"),
		MainDelegate:           new(MainDelegateDummy),
		LogLevel:               int(zapcore.DebugLevel),
		DocumentAudioDirectory: "./_files/audio",
		DocumentVideoDirectory: "./_files/video",
		DocumentPhotoDirectory: "./_files/photo",
		DocumentFileDirectory:  "./_files/file",
		DocumentCacheDirectory: "./_files/cache",
		DocumentLogDirectory:   "./_files/logs",
		ConnInfo:               conInfo,
	})
	_River = r
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
