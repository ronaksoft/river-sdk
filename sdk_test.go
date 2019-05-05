package riversdk

import (
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
	"testing"
)

var (
	_River   *River
	wg       *sync.WaitGroup
	testCase int
	test     *testing.T
)

func init() {
	logs.Info("Creating New River SDK Instance")
	r := new(River)
	conInfo := new(RiverConnection)
	conInfo.Delegate = new(dummyConInfoDelegate)
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
		ConnInfo:               conInfo,
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

func TestGetWorkGroup(t *testing.T) {
	b, err := GetWorkGroup("ws://new.river.im", 15)
	if err != nil {
		t.Error(err)
		return
	}
	si := new(msg.SystemInfo)
	err = si.Unmarshal(b)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("WorkGroupName:", si.WorkGroupName)
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
	_River = r
}

func TestRiver_SearchGlobal(t *testing.T) {
	var nonContactWithDialogUser = "nonContactWithDialogUser"
	var nonContactWhitoutDialogUser = "nonContactWithoutDialogUser"
	var ContactUser = "contactUser"
	var groupTitle = "groupTitle"
	message := new(msg.UserMessage)
	message.PeerID = 123
	message.ID = 123
	message.PeerType = 1
	message.Body = "Collectors are used to gather information about the system. By default a set of collectors is activated. You can see the details about the set in the README-file. If you want to use a specific set of collectors, you can define them in the ExecStart section of the service. Collectors are enabled by providing a--collector.<name> flag. Collectors that are enabled by default can be disabled by providing a --no-collector.<name> flag ، مجموعه اسپا و تندرستی حس خوب زندگی با کسب امتیاز در 5 شاخص از میان 11 محور ارزیابی شده، در میان بیش از 25000 باشگاه و مجموعه ورزشی کل کشور، بالاترین میزان رشد و عملکرد سازنده را به خود اختصاص داد."
	nonContactWithDialog := new(msg.User)
	nonContactWithDialog.ID = 321
	nonContactWithDialog.Username = nonContactWithDialogUser
	_ = repo.Users.SaveUser(nonContactWithDialog)

	nonContactWithoutDialog := new(msg.User)
	nonContactWithoutDialog.ID = 654
	nonContactWithoutDialog.Username = nonContactWhitoutDialogUser
	_ = repo.Users.SaveUser(nonContactWithoutDialog)

	contact := new(msg.ContactUser)
	contact.ID = 852
	contact.AccessHash = 4548
	contact.Username = ContactUser
	_ = repo.Users.SaveContactUser(contact)

	dialog := new(msg.Dialog)
	dialog.PeerType = 1
	dialog.PeerID = 321
	_ = repo.Dialogs.SaveDialog(dialog, 0)
	group := new(msg.Group)
	group.ID = 987
	group.Title = groupTitle
	_ = repo.Groups.Save(group)

	_ = repo.Messages.SaveMessage(message)
	wg = new(sync.WaitGroup)
	test = t
	wg.Add(1)
	testCase = 1
	_River.SearchGlobal("information about")
	wg.Wait()

	wg.Add(1)
	testCase = 2
	_River.SearchGlobal(ContactUser)
	wg.Wait()

	wg.Add(1)
	testCase = 3
	_River.SearchGlobal(nonContactWithDialogUser)
	wg.Wait()

	wg.Add(1)
	testCase = 4
	_River.SearchGlobal(nonContactWhitoutDialogUser)
	wg.Wait()
}

func (d *MainDelegateDummy) OnSearchComplete(b []byte) {
	logs.Info("OnSearchComplete")
	result := new(msg.ClientSearchResult)
	err := result.Unmarshal(b)
	if err != nil {
		test.Error("error Unmarshal", zap.String("", err.Error()))
		//logs.Warn("error Unmarshal", zap.String("", err.Error()))
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

type dummyConInfoDelegate struct{}

func (c *dummyConInfoDelegate) SaveConnInfo(connInfo []byte) {

}
