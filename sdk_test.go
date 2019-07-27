package riversdk

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"git.ronaksoftware.com/ronak/riversdk/pkg/salt"
	"git.ronaksoftware.com/ronak/toolbox"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

var (
	_River *River
)

var ConnInfo []byte

type ConnInfoDelegates struct{}

func (c *ConnInfoDelegates) SaveConnInfo(connInfo []byte) {
	_ = os.MkdirAll("./_connection", os.ModePerm)
	ConnInfo = connInfo
	err := ioutil.WriteFile("./_connection/connInfo", connInfo, 0666)
	if err != nil {
		fmt.Println(err)
	}
}

func init() {
	// logs.Info("Creating New River SDK Instance")
	// r := new(River)
	// conInfo := new(RiverConnection)
	// conInfo.Delegate = new(dummyConInfoDelegate)
	// r.SetConfig(&RiverConfig{
	// 	DbPath:                 "./_data/",
	// 	DbID:                   "test",
	// 	ServerKeysFilePath:     "./keys.json",
	// 	WebsocketEndpoint:         "ws://test.river.im",
	// 	QueuePath:              fmt.Sprintf("%s/%s", "./_queue", "test"),
	// 	MainDelegate:           new(MainDelegateDummy),
	// 	LogLevel:               int(zapcore.DebugLevel),
	// 	DocumentAudioDirectory: "./_files/audio",
	// 	DocumentVideoDirectory: "./_files/video",
	// 	DocumentPhotoDirectory: "./_files/photo",
	// 	DocumentFileDirectory:  "./_files/file",
	// 	DocumentCacheDirectory: "./_files/cache",
	// 	DocumentLogDirectory:   "./_files/logs",
	// 	ConnInfo:               conInfo,
	// })
	// _River = r
}

func TestController_CheckSalt(t *testing.T) {
	_ = repo.InitRepo(fmt.Sprintf("%s/%s.db", "./_data", "test"), false)
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
			err := repo.System.SaveString(domain.ColumnSystemSalts, string(b))
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
}

func TestGetWorkGroup(t *testing.T) {
	b, err := GetWorkGroup("ws://alaki.river.im", 5)
	if err != nil {
		t.Error(err)
		time.Sleep(10 * time.Second)
		return
	}
	si := new(msg.SystemInfo)
	err = si.Unmarshal(b)
	if err != nil {
		t.Error(err)
		time.Sleep(10 * time.Second)
		return
	}
	t.Log("WorkGroupName:", si.WorkGroupName)

}

func TestSDKReconnect(t *testing.T) {
	logs.SetLogLevel(0)
	fmt.Println("Creating New River SDK Instance")
	conInfo := new(RiverConnection)

	file, err := os.Open("./_connection/connInfo1")
	if err == nil {
		b, _ := ioutil.ReadAll(file)
		err := json.Unmarshal(b, conInfo)
		if err != nil {
			t.Error(err)
			return
		}
	}

	conInfo.Delegate = new(ConnInfoDelegates)

	r := new(River)
	fmt.Println("SetConfig called")
	r.SetConfig(&RiverConfig{
		DbPath:             "./_data/",
		DbID:               "test",
		QueuePath:          "./_queue/",
		ServerKeysFilePath: "./keys.json",
		ServerEndpoint:     "ws://new.river.im",
		LogLevel:           0,
		ConnInfo:           conInfo,
	})

	fmt.Println("Start called")
	_ = r.Start()
	for r.ConnInfo.AuthID == 0 {
		if err := r.CreateAuthKey(); err != nil {
			t.Error(err.Error())
			return
		}
	}

	time.Sleep(10 * time.Second)
	r.ResetAuthKey()
	r.Stop()

	time.Sleep(10 * time.Second)

	// Connect to 2nd Server
	file, err = os.Open("./_connection/connInfo2")
	if err == nil {
		b, _ := ioutil.ReadAll(file)
		err := json.Unmarshal(b, conInfo)
		if err != nil {
			t.Error(err)
			return
		}
	}

	conInfo.Delegate = new(ConnInfoDelegates)

	r.SetConfig(&RiverConfig{
		DbPath:             "./_data/",
		DbID:               "test",
		QueuePath:          "./_queue/",
		ServerKeysFilePath: "./keys.json",
		ServerEndpoint:     "ws://new.river.im",
		ConnInfo:           conInfo,
		LogLevel:           0,
	})
	_ = r.Start()
	for r.ConnInfo.AuthID == 0 {
		logs.Info("AuthKey has not been created yet.")
		if err := r.CreateAuthKey(); err != nil {
			t.Error(err.Error())
			return
		}
		logs.Info("AuthKey Created.")
	}
	fmt.Println("AuthID", r.ConnInfo.AuthID)
	fmt.Println("AuthKey", r.ConnInfo.AuthKey)
	time.Sleep(time.Second * 10)
	b := r.GetMonitorStats()
	fmt.Println(string(b))

}

func TestConnectTime(t *testing.T) {
	fmt.Println("Creating New River SDK Instance")
	conInfo := new(RiverConnection)

	file, err := os.Open("./_connection/connInfo1")
	if err == nil {
		b, _ := ioutil.ReadAll(file)
		err := json.Unmarshal(b, conInfo)
		if err != nil {
			t.Error(err)
			return
		}
	}

	conInfo.Delegate = new(ConnInfoDelegates)
	conInfo.UserID = 10
	conInfo.AuthID = -309683086834753830
	authKey, _ := base64.StdEncoding.DecodeString(
		"w5lswq0Vw6Eqwq/DrMKSRTHDu2rDsMOmE1XDuXLCgsKiwqVfwp7Cpk7CrTUgexXDlALDl1bCi8OSBcOYw7AIw4jCvsK9wrkgwqjDvsOpwrTDtMKQwqA9MEPCg2Q6cR8" +
			"+Qi1hw6fCmsK9RcK3UsOifcKAPcOZw4zDtksbwohGF8OiTMOGw6pRe8K/AsKTwqdsTsO/aRINw7/DnCvDumrDtgc3RcKHcV5UwqLDs8OTDsK" +
			"+csODXyPCu8KQw4UjHEnDjcO2b8KRwrlCw57Cs8KCIsOlwqdcS0PCr8OGwrnCgAvCu0TCisOeRF8nwrNqIF3CqyMkw7HCi8KVF2zDlG5JVsKuf8KYMyLClHbCvcOGED5rwqVTLi1NVcOOLcKjHsKqw51jdB1xKMOoGcKBEMOeJcO4VgtAw6LCt2MRVznDqMOawpjDgcOvwq7CiHJDwrrCpcO4OcKOwpEwJxITwp8Xw5guwoJ7fMOJTl3Ctz0CXjQlwpHCrBTDvQES")
	copy(conInfo.AuthKey[:], authKey)

	r := new(River)
	fmt.Println("SetConfig called")
	r.SetConfig(&RiverConfig{
		DbPath:             "./_data/",
		DbID:               "test",
		QueuePath:          "./_queue/",
		ServerKeysFilePath: "./keys.json",
		ServerEndpoint:     "ws://new.river.im",
		LogLevel:           0,
		ConnInfo:           conInfo,
	})

	fmt.Println("Start called")
	_ = r.Start()
	for r.ConnInfo.AuthID == 0 {
		if err := r.CreateAuthKey(); err != nil {
			t.Error(err.Error())
			return
		}
	}

	time.Sleep(time.Second * 10)
	b := r.GetMonitorStats()
	fmt.Println(string(b))

}
func TestNewRiver(t *testing.T) {
	logs.Info("Creating New River SDK Instance")
	conInfo := new(RiverConnection)

	file, err := os.Open("./_connection/connInfo1")
	if err == nil {
		b, _ := ioutil.ReadAll(file)
		err := json.Unmarshal(b, conInfo)
		if err != nil {
			t.Error(err)
			return
		}
	}

	conInfo.Delegate = new(ConnInfoDelegates)

	r := new(River)
	r.SetConfig(&RiverConfig{
		DbPath:             "./_data/",
		DbID:               "test",
		QueuePath:          "./_queue/",
		ServerKeysFilePath: "./keys.json",
		ServerEndpoint:     "ws://new.river.im",
		LogLevel:           0,
		ConnInfo:           conInfo,
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

	updateGetState := new(msg.UpdateGetState)
	b, _ := updateGetState.Marshal()

	for i := 0; i < 10; i++ {
		reqID, err := r.ExecuteCommand(msg.C_UpdateGetState, b, new(RequestDelegateDummy), true, true)
		if err != nil {
			t.Error(reqID, ":::", err)
			return
		}
		t.Log("RequestID:", reqID)
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

type RequestDelegateDummy struct{}

func (RequestDelegateDummy) OnComplete(b []byte) {
	fmt.Println(b)
}

func (RequestDelegateDummy) OnTimeout(err error) {
	fmt.Println(err)
}

type FileDelegateDummy struct{}

func (d *FileDelegateDummy) OnDownloadProgressChanged(messageID, processedParts, totalParts int64, percent float64) {
	logs.Info("Download progress changed", zap.Float64("Progress", percent))
}

func (d *FileDelegateDummy) OnUploadProgressChanged(messageID, processedParts, totalParts int64, percent float64) {
	logs.Info("Upload progress changed", zap.Float64("Progress", percent))
}

func (d *FileDelegateDummy) OnDownloadCompleted(messageID int64, filePath string) {
	logs.Info("Download completed", zap.Int64("MsgID", messageID), zap.String("FilePath", filePath))
}

func (d *FileDelegateDummy) OnUploadCompleted(messageID int64, filePath string) {
	logs.Info("Upload completed", zap.Int64("MsgID", messageID), zap.String("FilePath", filePath))
}

func (d *FileDelegateDummy) OnUploadError(messageID, requestID int64, filePath string, err []byte) {
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

func (d *FileDelegateDummy) OnDownloadError(messageID, requestID int64, filePath string, err []byte) {
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

//func (d *MainDelegateDummy) OnSearchComplete(b []byte) {
//	logs.Info("OnSearchComplete")
//	result := new(msg.ClientSearchResult)
//	err := result.Unmarshal(b)
//	if err != nil {
//		test.Error("error Unmarshal", zap.String("", err.Error()))
//		return
//	}
//	switch testCase {
//	case 1:
//		if len(result.Messages) > 0 {
//			if result.Messages[0].ID != 123 {
//				test.Error(fmt.Sprintf("expected msg ID 123, have %d", result.Messages[0].ID))
//			}
//		} else {
//			test.Error(fmt.Sprintf("expected msg ID 123, have not any"))
//		}
//
//		wg.Done()
//	case 2:
//		if len(result.Messages) > 0 {
//			test.Error(fmt.Sprintf("expected no messages"))
//		}
//		if len(result.MatchedUsers) > 0 {
//			if result.MatchedUsers[0].ID != 852 {
//				test.Error(fmt.Sprintf("expected user ID 852, have %d", result.Messages[0].ID))
//			}
//		} else {
//			test.Error(fmt.Sprintf("expected user ID 852, have nothing, %+v", result))
//		}
//		wg.Done()
//	case 3:
//		if len(result.Messages) > 0 {
//			test.Error(fmt.Sprintf("expected no messages"))
//		}
//		if len(result.MatchedUsers) > 0 {
//			if result.MatchedUsers[0].ID != 321 {
//				test.Error(fmt.Sprintf("expected user ID 321, have %d", result.Messages[0].ID))
//			}
//		} else {
//			test.Error(fmt.Sprintf("expected user ID 321, have nothing, %+v", result))
//		}
//		wg.Done()
//	case 4:
//		if len(result.Messages) > 0 || len(result.MatchedUsers) > 0 || len(result.MatchedGroups) > 0 {
//			test.Error(fmt.Sprintf("expected to found nothing but found %v", result))
//		}
//		wg.Done()
//	}
//}
