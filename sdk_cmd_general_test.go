package riversdk

import (
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
	"testing"
)

var (
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
		DbPath: "./_data/",
		DbID:   "test",
		// ServerKeysFilePath:     "./keys.json",
		ServerEndpoint:         "ws://new.river.im",
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
	})
	_River = r
}

func TestRiver_SearchGlobal(t *testing.T) {
	var nonContactWithDialogUser = "nonContactWithDialogUser"
	var nonContactWhitoutDialogUser = "nonContactWithoutDialogUser"
	var ContactUser = "contactUser"
	var groupTitle = "groupTitle"
	createDataForSearchGlobal(nonContactWithDialogUser, nonContactWhitoutDialogUser, ContactUser, groupTitle)
	searchDelegate := new(searchGlobalDelegateDummy)
	tests := []struct {
		name       string
		searchText string
		peerID     int64
	}{
		{"search message body", "information about", 0},
		{"search contact name", ContactUser, 0},
		{"search non contact name who has dialog with user", nonContactWithDialogUser, 0},
		{"search non contact name who has NOT dialog with user", nonContactWhitoutDialogUser, 0},
	}
	wg = new(sync.WaitGroup)
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wg.Add(1)
			testCase = i + 1
			_River.SearchGlobal(tt.searchText, tt.peerID, searchDelegate)
			wg.Wait()
		})
	}
}

type searchGlobalDelegateDummy struct{}

func (searchGlobalDelegateDummy) OnComplete(b []byte) {
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
		logs.Debug("Result", zap.Any("", result))

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

func (searchGlobalDelegateDummy) OnTimeout(err error) {
	fmt.Println(err)
}

type dummyConInfoDelegate struct{}

func (c *dummyConInfoDelegate) SaveConnInfo(connInfo []byte) {}

func createDataForSearchGlobal(nonContactWithDialogUser, nonContactWhitoutDialogUser, ContactUser, groupTitle string) {
	message := new(msg.UserMessage)
	message.PeerID = 123
	message.ID = 123
	message.PeerType = 1
	message.Body = "Collectors are used to gather information about the system. By default a set of collectors is activated. You can see the details about the set in the README-file. If you want to use a specific set of collectors, you can define them in the ExecStart section of the service. Collectors are enabled by providing a--collector.<name> flag. Collectors that are enabled by default can be disabled by providing a --no-collector.<name> flag ، مجموعه اسپا و تندرستی حس خوب زندگی با کسب امتیاز در 5 شاخص از میان 11 محور ارزیابی شده، در میان بیش از 25000 باشگاه و مجموعه ورزشی کل کشور، بالاترین میزان رشد و عملکرد سازنده را به خود اختصاص داد."
	nonContactWithDialog := new(msg.User)
	nonContactWithDialog.ID = 321
	nonContactWithDialog.Username = nonContactWithDialogUser
	repo.Users.Save(nonContactWithDialog)

	nonContactWithoutDialog := new(msg.User)
	nonContactWithoutDialog.ID = 654
	nonContactWithoutDialog.Username = nonContactWhitoutDialogUser
	repo.Users.Save(nonContactWithoutDialog)

	contact := new(msg.ContactUser)
	contact.ID = 852
	contact.AccessHash = 4548
	contact.Username = ContactUser
	repo.Users.SaveContact(contact)

	dialog := new(msg.Dialog)
	dialog.PeerType = 1
	dialog.PeerID = 321
	repo.Dialogs.Save(dialog)
	group := new(msg.Group)
	group.ID = 987
	group.Title = groupTitle
	repo.Groups.Save(group)

}
