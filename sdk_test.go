package riversdk

import (
	"encoding/json"
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
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

func init() {}

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

func TestReconnect(t *testing.T) {
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
		LogLevel:           -1,
		ConnInfo:           conInfo,
	})

	r.Start()
	for r.ConnInfo.AuthID == 0 {
		logs.Info("AuthKey has not been created yet.")
		if err := r.CreateAuthKey(); err != nil {
			t.Error(err.Error())
			return
		}
		logs.Info("AuthKey Created.")
	}


	time.Sleep(10 * time.Second)
	r.Stop()
	r.ResetAuthKey()

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
		ServerEndpoint:     "ws://test.river.im",
		ConnInfo:           conInfo,
		LogLevel: -1,
	})
	r.Start()
	for r.ConnInfo.AuthID == 0 {
		logs.Info("AuthKey has not been created yet.")
		if err := r.CreateAuthKey(); err != nil {
			t.Error(err.Error())
			return
		}
		logs.Info("AuthKey Created.")
	}
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
	b,_ := updateGetState.Marshal()

	for i := 0;i < 10 ;i++ {
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


type RequestDelegateDummy struct {}

func (RequestDelegateDummy) OnComplete(b []byte) {
	fmt.Println(b)
}

func (RequestDelegateDummy) OnTimeout(err error) {
	fmt.Println(err)
}

func (d *MainDelegateDummy) OnSearchComplete(b []byte) {
	logs.Info("OnSearchComplete")
	result := new(msg.ClientSearchResult)
	err := result.Unmarshal(b)
	if err != nil {
		test.Error("error Unmarshal", zap.String("", err.Error()))
		return
	}
	switch testCase {
	case 1:
		if len(result.Messages) > 0 {
			if result.Messages[0].ID != 123 {
				test.Error(fmt.Sprintf("expected msg ID 123, have %d", result.Messages[0].ID))
			}
		} else {
			test.Error(fmt.Sprintf("expected msg ID 123, have not any"))
		}

		wg.Done()
	case 2:
		if len(result.Messages) > 0 {
			test.Error(fmt.Sprintf("expected no messages"))
		}
		if len(result.MatchedUsers) > 0 {
			if result.MatchedUsers[0].ID != 852 {
				test.Error(fmt.Sprintf("expected user ID 852, have %d", result.Messages[0].ID))
			}
		} else {
			test.Error(fmt.Sprintf("expected user ID 852, have nothing, %+v", result))
		}
		wg.Done()
	case 3:
		if len(result.Messages) > 0 {
			test.Error(fmt.Sprintf("expected no messages"))
		}
		if len(result.MatchedUsers) > 0 {
			if result.MatchedUsers[0].ID != 321 {
				test.Error(fmt.Sprintf("expected user ID 321, have %d", result.Messages[0].ID))
			}
		} else {
			test.Error(fmt.Sprintf("expected user ID 321, have nothing, %+v", result))
		}
		wg.Done()
	case 4:
		if len(result.Messages) > 0 || len(result.MatchedUsers) > 0 || len(result.MatchedGroups) > 0 {
			test.Error(fmt.Sprintf("expected to found nothing but found %v", result))
		}
		wg.Done()
	}
}
