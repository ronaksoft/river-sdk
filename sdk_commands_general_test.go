package riversdk

import (
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
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
	_River = r
}

func TestRiver_SearchGlobal(t *testing.T) {
	var nonContactWithDialogUser = "nonContactWithDialogUser"
	var nonContactWhitoutDialogUser = "nonContactWithoutDialogUser"
	var ContactUser = "contactUser"
	var groupTitle = "groupTitle"
	createDataForSearchGlobal(nonContactWithDialogUser, nonContactWhitoutDialogUser, ContactUser, groupTitle)
	tests := []struct {
		name       string
		searchText string
	}{
		{"search message body", "information about"},
		{"search contact name", ContactUser},
		{"search non contact name who has dialog with user", nonContactWithDialogUser},
		{"search non contact name who has NOT dialog with user", nonContactWhitoutDialogUser},
	}
	wg = new(sync.WaitGroup)
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wg.Add(1)
			testCase = i + 1
			_River.SearchGlobal(tt.searchText, nil)
			wg.Wait()
		})
	}
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
}
