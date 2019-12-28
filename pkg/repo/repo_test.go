package repo_test

import (
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/chat"
	fileCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_file"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"
	"sync"
	"testing"
	"time"
)

/*
   Creation Time: 2019 - Jul - 20
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software GroupSearch 2018
*/

func init() {
	repo.InitRepo("./_data", false)
}

func createMediaMessage(body string, filename string, labelIDs []int32) *msg.UserMessage {
	userID := ronak.RandomInt64(0)
	attrFile, _ := (&msg.DocumentAttributeFile{Filename: filename}).Marshal()
	media, _ := (&msg.MediaDocument{
		Caption:      "This is caption",
		TTLinSeconds: 0,
		Doc: &msg.Document{
			ID:         ronak.RandomInt64(0),
			AccessHash: ronak.RandomUint64(),
			Date:       time.Now().Unix(),
			MimeType:   "",
			FileSize:   1243,
			Version:    0,
			ClusterID:  1,
			Attributes: []*msg.DocumentAttribute{
				{
					Type: msg.AttributeTypeFile,
					Data: attrFile,
				},
			},
			Thumbnail:   nil,
			MD5Checksum: "",
		},
	}).Marshal()
	return &msg.UserMessage{
		ID:                  userID,
		PeerID:              ronak.RandomInt64(0),
		PeerType:            1,
		CreatedOn:           time.Now().Unix(),
		EditedOn:            0,
		FwdSenderID:         0,
		FwdChannelID:        0,
		FwdChannelMessageID: 0,
		Flags:               0,
		MessageType:         0,
		Body:                body,
		SenderID:            userID,
		ContentRead:         false,
		Inbox:               false,
		ReplyTo:             0,
		MessageAction:       0,
		MessageActionData:   nil,
		Entities:            nil,
		MediaType:           msg.MediaTypeDocument,
		Media:               media,
		ReplyMarkup:         0,
		ReplyMarkupData:     nil,
		LabelIDs:            labelIDs,
	}
}
func createMessage(id int64, body string, labelIDs []int32) *msg.UserMessage {
	userID := ronak.RandomInt64(0)

	return &msg.UserMessage{
		ID:                  id,
		PeerID:              ronak.RandomInt64(0),
		PeerType:            1,
		CreatedOn:           time.Now().Unix(),
		EditedOn:            0,
		FwdSenderID:         0,
		FwdChannelID:        0,
		FwdChannelMessageID: 0,
		Flags:               0,
		MessageType:         0,
		Body:                body,
		SenderID:            userID,
		ContentRead:         false,
		Inbox:               false,
		ReplyTo:             0,
		MessageAction:       0,
		MessageActionData:   nil,
		Entities:            nil,
		MediaType:           msg.MediaTypeEmpty,
		Media:               nil,
		ReplyMarkup:         0,
		ReplyMarkupData:     nil,
		LabelIDs:            labelIDs,
	}
}
func TestRepoDialogs(t *testing.T) {
	dialog := new(msg.Dialog)
	dialog.PeerID = 100
	dialog.PeerType = 1
	dialog.TopMessageID = 1000
	dialog.ReadOutboxMaxID = 900
	dialog.ReadInboxMaxID = 901
	// err := repo.Dialogs.save(dialog)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	repo.Dialogs.SaveNew(dialog, time.Now().Unix())

	d, _ := repo.Dialogs.Get(100, 1)
	t.Log(dialog)
	t.Log(d)
}

func TestRepoMessagesExtra(t *testing.T) {
	repo.MessagesExtra.SaveScrollID(11, 1, 101)
	x := repo.MessagesExtra.GetScrollID(11, 1)
	fmt.Println(x)
}

func TestPending(t *testing.T) {
	pm := new(msg.ClientSendMessageMedia)
	pm.Peer = new(msg.InputPeer)
	_, err := repo.PendingMessages.SaveClientMessageMedia(10, 1, 11, 20, 21, pm)
	if err != nil {
		t.Error(err)
	}
	pm1 := repo.PendingMessages.GetByID(10)
	fmt.Println(pm1)

	repo.PendingMessages.Delete(10)

	pm2 := repo.PendingMessages.GetByID(10)
	fmt.Println(pm2)
}

func TestGetMessageKey(t *testing.T) {
	peerID := 10001
	peerType := 1
	msgID := 1 << 32
	ronak.StrToByte(fmt.Sprintf("%s.%021d.%d.%012d", "MSG", peerID, peerType, msgID))
	fmt.Println(fmt.Sprintf("%s.%021d.%d.%012d", "MSG", peerID, peerType, msgID))
}

func TestRepoDeleteMessage(t *testing.T) {
	peerID := int64(10001)
	peerType := int32(1)

	d, _ := repo.Dialogs.Get(peerID, peerType)
	if d == nil {
		d = new(msg.Dialog)
		d.PeerID = peerID
		d.PeerType = peerType
		repo.Dialogs.Save(d)
	}

	for i := int64(10); i < 20; i++ {
		m := new(msg.UserMessage)
		m.ID = i
		m.PeerID = peerID
		m.PeerType = peerType
		m.SenderID = peerID
		m.Body = fmt.Sprintf("Text %d", i)
		repo.Messages.SaveNew(m, 10002)
	}

	d, _ = repo.Dialogs.Get(peerID, peerType)
	fmt.Println(d)

	repo.Messages.Delete(10002, peerID, peerType, 19)
	d, _ = repo.Dialogs.Get(peerID, peerType)
	fmt.Println(d)

	msgs, _ := repo.Messages.GetMessageHistory(peerID, peerType, 0, 0, 5)
	for idx := range msgs {
		fmt.Println(msgs[idx].ID)
	}

}

func TestConcurrent(t *testing.T) {
	waitGroup := sync.WaitGroup{}
	for i := int64(1); i < 10000; i++ {
		waitGroup.Add(1)
		go func(i int64) {
			_, err := repo.PendingMessages.SaveMessageMedia(i, 1001, &msg.MessagesSendMedia{
				RandomID: ronak.RandomInt64(0),
				Peer: &msg.InputPeer{
					ID:         i,
					Type:       msg.PeerUser,
					AccessHash: 0,
				},
				MediaType:  0,
				MediaData:  nil,
				ReplyTo:    0,
				ClearDraft: false,
			})
			waitGroup.Done()
			if err != nil {
				logs.Fatal("Error On Save Pending", zap.Error(err))
			}
		}(i)
		waitGroup.Add(1)
		go func(i int64) {
			err := repo.PendingMessages.Delete(i)
			waitGroup.Done()
			if err != nil {
				logs.Fatal("Error On Save Pending", zap.Error(err))
			}
		}(i)
	}
	waitGroup.Wait()

}

func TestClearHistory(t *testing.T) {
	m := make([]*msg.UserMessage, 0, 10)
	for i := 1; i < 1000; i++ {
		m = append(m, &msg.UserMessage{
			ID:                  int64(i),
			PeerID:              10,
			PeerType:            1,
			CreatedOn:           time.Now().Unix(),
			EditedOn:            0,
			FwdSenderID:         0,
			FwdChannelID:        0,
			FwdChannelMessageID: 0,
			Flags:               0,
			MessageType:         0,
			Body:                fmt.Sprintf("Hello %d", i),
			SenderID:            100,
			ContentRead:         false,
			Inbox:               false,
			ReplyTo:             0,
			MessageAction:       0,
			MessageActionData:   nil,
			Entities:            nil,
			MediaType:           0,
			Media:               nil,
		})
	}
	err := repo.Dialogs.Save(&msg.Dialog{
		PeerID:          10,
		PeerType:        1,
		TopMessageID:    999,
		ReadInboxMaxID:  0,
		ReadOutboxMaxID: 0,
		UnreadCount:     0,
		AccessHash:      0,
		NotifySettings:  nil,
		MentionedCount:  0,
		Pinned:          false,
		Draft:           nil,
	})
	if err != nil {
		t.Error(err)
		return
	}
	repo.Messages.Save(m...)
	fmt.Println("Saved")
	err = repo.Messages.ClearHistory(101, 10, 1, 995)
	if err != nil {
		t.Error(err)
		return
	}
	d, err := repo.Dialogs.Get(10, 1)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(d.TopMessageID)
	ums, us := repo.Messages.GetMessageHistory(10, 1, 0, 0, 100)
	fmt.Println(len(ums), len(us))

	var x []int64
	for _, um := range ums {
		x = append(x, um.ID)
	}
	fmt.Println(x)

	repo.Messages.Delete(101, 10, 1, 1950)
	d, err = repo.Dialogs.Get(10, 1)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(d.TopMessageID)
}

func TestSearch(t *testing.T) {
	m := make([]*msg.UserMessage, 0, 10)
	for i := 1; i < 100; i++ {
		peerID := int64(i%10 + 1)
		peerType := int32(msg.PeerUser)
		if i%2 == 0 {
			peerID = -peerID
			peerType = int32(msg.PeerGroup)
		}
		m = append(m, &msg.UserMessage{
			ID:                  int64(i),
			PeerID:              peerID,
			PeerType:            peerType,
			CreatedOn:           time.Now().Unix(),
			EditedOn:            0,
			FwdSenderID:         0,
			FwdChannelID:        0,
			FwdChannelMessageID: 0,
			Flags:               0,
			MessageType:         0,
			Body:                fmt.Sprintf("Hello %d %d", i, peerType),
			SenderID:            100,
			ContentRead:         false,
			Inbox:               false,
			ReplyTo:             0,
			MessageAction:       0,
			MessageActionData:   nil,
			Entities:            nil,
			MediaType:           0,
			Media:               nil,
		})
	}
	repo.Messages.Save(m...)
	fmt.Println("Saved")

	// mm := repo.Messages.SearchText("Hello")
	fmt.Print("Search in UserPeer:")
	mm := repo.Messages.SearchTextByPeerID("H", 6, 100)
	for _, m := range mm {
		fmt.Println(m.ID, m.Body, m.PeerID)
	}
	fmt.Print("Search in GroupPeer:")
	mm = repo.Messages.SearchTextByPeerID("H", -7, 100)
	for _, m := range mm {
		fmt.Println(m.ID, m.Body, m.PeerID)
	}
}

func TestUserPhotoGallery(t *testing.T) {
	photo1 := &msg.UserPhoto{
		PhotoBig: &msg.FileLocation{
			ClusterID:  100,
			FileID:     200,
			AccessHash: 300,
		},
		PhotoSmall: &msg.FileLocation{
			ClusterID:  10,
			FileID:     20,
			AccessHash: 30,
		},
		PhotoID: 1,
	}
	photo2 := &msg.UserPhoto{
		PhotoBig: &msg.FileLocation{
			ClusterID:  101,
			FileID:     201,
			AccessHash: 301,
		},
		PhotoSmall: &msg.FileLocation{
			ClusterID:  11,
			FileID:     21,
			AccessHash: 31,
		},
		PhotoID: 2,
	}
	user := &msg.User{
		ID:           1000,
		FirstName:    "Ehsan",
		LastName:     "Noureddin Moosa",
		Username:     "",
		Status:       0,
		Restricted:   false,
		AccessHash:   0,
		Photo:        photo1,
		Bio:          "",
		Phone:        "",
		LastSeen:     0,
		PhotoGallery: nil,
		IsBot:        false,
	}
	repo.Users.Save(user)

	u1, err := repo.Users.Get(1000)
	if err != nil {
		t.Fatal(err)
	}

	user.PhotoGallery = []*msg.UserPhoto{photo1, photo2}
	repo.Users.Save(user)
	u2, err := repo.Users.Get(1000)
	if err != nil {
		t.Fatal(err)
	}

	phGallery := repo.Users.GetPhotoGallery(1000)
	fmt.Println(phGallery)
	_ = u1
	_ = u2
}

func TestGroupPhotoGallery(t *testing.T) {
	photo1 := &msg.GroupPhoto{
		PhotoBig: &msg.FileLocation{
			ClusterID:  100,
			FileID:     200,
			AccessHash: 300,
		},
		PhotoSmall: &msg.FileLocation{
			ClusterID:  10,
			FileID:     20,
			AccessHash: 30,
		},
		PhotoID: 1,
	}
	photo2 := &msg.GroupPhoto{
		PhotoBig: &msg.FileLocation{
			ClusterID:  101,
			FileID:     201,
			AccessHash: 301,
		},
		PhotoSmall: &msg.FileLocation{
			ClusterID:  11,
			FileID:     21,
			AccessHash: 31,
		},
		PhotoID: 2,
	}
	group := &msg.GroupFull{
		Group: &msg.Group{
			ID:           1000,
			Title:        "Test Group",
			CreatedOn:    0,
			Participants: 0,
			EditedOn:     0,
			Flags:        nil,
			Photo:        photo1,
		},
		Users:          nil,
		Participants:   nil,
		NotifySettings: nil,
		PhotoGallery:   []*msg.GroupPhoto{photo1, photo2},
	}

	repo.Groups.Save(group.Group)
	repo.Groups.SavePhotoGallery(group.Group.ID, group.PhotoGallery...)

	phGallery := repo.Groups.GetPhotoGallery(1000)
	fmt.Println(phGallery)
}

func TestMessagesSave(t *testing.T) {
	m := createMediaMessage("Hello", "file.txt", nil)
	repo.Messages.Save(m)
	media := &msg.MediaDocument{}
	media.Unmarshal(m.Media)
	clientFile, err := repo.Files.Get(media.Doc.ClusterID, media.Doc.ID, media.Doc.AccessHash)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(clientFile.Extension)
	t.Log(fileCtrl.GetFilePath(clientFile))
}

func TestLabel(t *testing.T) {
	Convey("Testing Labels", t, func(c C) {
		Convey("Save Labels", func(c C) {
			err := repo.Labels.Save(
				&msg.Label{
					ID:     1,
					Name:   "Label1",
					Colour: "#FF0000",
				},
				&msg.Label{
					ID:     2,
					Name:   "Label 2",
					Colour: "#00FF00",
				},
			)
			c.So(err, ShouldBeNil)
			labels := repo.Labels.GetAll()
			c.So(labels, ShouldHaveLength, 2)
		})
		Convey("Add Label To Message", func(c C) {
			peerID := int64(100)
			for i := 1; i <= 10; i++ {
				repo.Messages.Save(&msg.UserMessage{
					ID:                  int64(i),
					PeerID:              peerID,
					PeerType:            0,
					CreatedOn:           0,
					EditedOn:            0,
					FwdSenderID:         0,
					FwdChannelID:        0,
					FwdChannelMessageID: 0,
					Flags:               0,
					MessageType:         0,
					Body:                fmt.Sprintf("Test %d", i),
					SenderID:            0,
					ContentRead:         false,
					Inbox:               false,
					ReplyTo:             0,
					MessageAction:       0,
					MessageActionData:   nil,
					Entities:            nil,
					MediaType:           0,
					Media:               nil,
					ReplyMarkup:         0,
					ReplyMarkupData:     nil,
					LabelIDs:            nil,
				})
			}

			err := repo.Labels.AddLabelsToMessages([]int32{1}, 1, peerID, []int64{1, 2, 3, 6, 8, 9, 10})
			c.So(err, ShouldBeNil)
		})
		Convey("List Messages", func(c C) {
			ums, _, _ := repo.Labels.ListMessages(1, 3, 0, 0)
			c.So(ums, ShouldHaveLength, 3)
			c.So(ums[0].ID, ShouldEqual, 1)
			c.So(ums[1].ID, ShouldEqual, 2)
			c.So(ums[2].ID, ShouldEqual, 3)

			ums, _, _ = repo.Labels.ListMessages(1, 2, 6, 0)
			c.So(ums, ShouldHaveLength, 2)
			c.So(ums[0].ID, ShouldEqual, 8)
			c.So(ums[1].ID, ShouldEqual, 6)

			ums, _, _ = repo.Labels.ListMessages(1, 3, 0, 9)
			c.So(ums, ShouldHaveLength, 3)
			c.So(ums[0].ID, ShouldEqual, 9)
			c.So(ums[1].ID, ShouldEqual, 8)
			c.So(ums[2].ID, ShouldEqual, 6)
		})

		Convey("Label Hole", func(c C) {
			b, _ := repo.Labels.GetLowerFilled(10, 100)
			c.So(b, ShouldBeFalse)

			err := repo.Labels.Fill(1, 10, 100)
			c.So(err, ShouldBeNil)

			bar := repo.Labels.GetFilled(1)
			c.So(bar.MinID, ShouldEqual, 10)
			c.So(bar.MaxID, ShouldEqual, 100)

			b, bar = repo.Labels.GetUpperFilled(1, 90)
			c.So(b, ShouldBeTrue)
			c.So(bar.MinID, ShouldEqual, 90)
			c.So(bar.MaxID, ShouldEqual, 100)

			b, bar = repo.Labels.GetLowerFilled(1, 90)
			c.So(b, ShouldBeTrue)
			c.So(bar.MinID, ShouldEqual, 10)
			c.So(bar.MaxID, ShouldEqual, 90)
		})

		Convey("Search By Label", func(c C) {

		})
	})
}
