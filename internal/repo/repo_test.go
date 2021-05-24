package repo_test

import (
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/testenv"
	"github.com/ronaksoft/rony/tools"
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
	repo.MustInit("./_data", false)
	testenv.Log().SetLogLevel(2)
}

func createMediaMessage(body string, filename string, labelIDs []int32) *msg.UserMessage {
	userID := domain.RandomInt63()
	attrFile, _ := (&msg.DocumentAttributeFile{Filename: filename}).Marshal()
	media, _ := (&msg.MediaDocument{
		Caption:      "This is caption",
		TTLinSeconds: 0,
		Doc: &msg.Document{
			ID:         domain.RandomInt63(),
			AccessHash: domain.RandomUint64(),
			Date:       time.Now().Unix(),
			MimeType:   "",
			FileSize:   1243,
			Version:    0,
			ClusterID:  1,
			Attributes: []*msg.DocumentAttribute{
				{
					Type: msg.DocumentAttributeType_AttributeTypeFile,
					Data: attrFile,
				},
			},
			Thumbnail:   nil,
			MD5Checksum: "",
		},
	}).Marshal()
	return &msg.UserMessage{
		ID:                  userID,
		PeerID:              domain.RandomInt63(),
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
		MediaType:           msg.MediaType_MediaTypeDocument,
		Media:               media,
		ReplyMarkup:         0,
		ReplyMarkupData:     nil,
		LabelIDs:            labelIDs,
	}
}
func createMessage(id int64, peerID int64, body string, labelIDs []int32) *msg.UserMessage {
	userID := domain.RandomInt63()

	return &msg.UserMessage{
		ID:                  id,
		PeerID:              peerID,
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
		MediaType:           msg.MediaType_MediaTypeEmpty,
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

	d, _ := repo.Dialogs.Get(0, 100, 1)
	t.Log(dialog)
	t.Log(d)
}

func TestRepoMessagesExtra(t *testing.T) {
	repo.MessagesExtra.SaveScrollID(0, 11, 1, 0, 101)
	x := repo.MessagesExtra.GetScrollID(0, 11, 1, 0)
	fmt.Println(x)
}

func TestPending(t *testing.T) {
	pm := new(msg.ClientSendMessageMedia)
	pm.Peer = new(msg.InputPeer)
	_, err := repo.PendingMessages.SaveClientMessageMedia(0, 0, 10, 1, 11, 20, 21, pm, nil)
	if err != nil {
		t.Error(err)
	}
	pm1, _ := repo.PendingMessages.GetByID(10)
	fmt.Println(pm1)

	_ = repo.PendingMessages.Delete(10)
	pm2, _ := repo.PendingMessages.GetByID(10)
	fmt.Println(pm2)
}

func TestRepoDeleteMessage(t *testing.T) {
	Convey("RepoDeleteMessage", t, func(c C) {
		userID := tools.RandomInt64(0)
		peerID := tools.RandomInt64(0)
		peerType := int32(1)

		d, _ := repo.Dialogs.Get(0, peerID, peerType)
		if d == nil {
			d = new(msg.Dialog)
			d.PeerID = peerID
			d.PeerType = peerType
			err := repo.Dialogs.Save(d)
			c.So(err, ShouldBeNil)
		}

		for i := int64(10); i < 20; i++ {
			m := new(msg.UserMessage)
			m.ID = i
			m.PeerID = peerID
			m.PeerType = peerType
			m.SenderID = peerID
			m.Body = fmt.Sprintf("Text %d", i)
			err := repo.Messages.SaveNew(m, userID)
			c.So(err, ShouldBeNil)
		}

		d, _ = repo.Dialogs.Get(0, peerID, peerType)
		c.So(d.TopMessageID, ShouldEqual, 19)

		repo.Messages.Delete(userID, 0, peerID, peerType, 19)
		d, _ = repo.Dialogs.Get(0, peerID, peerType)
		c.So(d.TopMessageID, ShouldEqual, 18)

		msgs, _, _ := repo.Messages.GetMessageHistory(0, peerID, peerType, 0, 0, 5)
		c.So(msgs, ShouldHaveLength, 5)
		c.So(msgs[0].ID, ShouldEqual, 18)
		c.So(msgs[4].ID, ShouldEqual, 14)
	})

}

func TestConcurrent(t *testing.T) {
	waitGroup := sync.WaitGroup{}
	for i := int64(1); i < 10000; i++ {
		waitGroup.Add(1)
		go func(i int64) {
			_, err := repo.PendingMessages.SaveMessageMedia(0, 0, i, 1001, &msg.MessagesSendMedia{
				RandomID: domain.RandomInt63(),
				Peer: &msg.InputPeer{
					ID:         i,
					Type:       msg.PeerType_PeerUser,
					AccessHash: 0,
				},
				MediaType:  0,
				MediaData:  nil,
				ReplyTo:    0,
				ClearDraft: false,
			})
			waitGroup.Done()
			if err != nil {
				testenv.Log().Fatal("Error On Save Pending", zap.Error(err))
			}
		}(i)
		waitGroup.Add(1)
		go func(i int64) {
			_ = repo.PendingMessages.Delete(i)
			waitGroup.Done()
		}(i)
	}
	waitGroup.Wait()

}

func TestClearHistory(t *testing.T) {
	Convey("ClearHistory", t, func(c C) {
		peerID := tools.RandomInt64(0)
		userID := tools.RandomInt64(0)
		dialog := &msg.Dialog{
			TeamID:         0,
			PeerID:         peerID,
			PeerType:       1,
			TopMessageID:   1,
			UnreadCount:    0,
			MentionedCount: 0,
			AccessHash:     0,
		}
		err := repo.Dialogs.SaveNew(dialog, tools.TimeUnix())
		c.So(err, ShouldBeNil)

		for i := 1; i < 1000; i++ {
			err := repo.Messages.SaveNew(&msg.UserMessage{
				ID:                  int64(i),
				PeerID:              peerID,
				PeerType:            1,
				CreatedOn:           time.Now().Unix(),
				EditedOn:            0,
				FwdSenderID:         0,
				FwdChannelID:        0,
				FwdChannelMessageID: 0,
				Flags:               0,
				MessageType:         0,
				Body:                fmt.Sprintf("Hello %d", i),
				SenderID:            peerID,
				ContentRead:         false,
				Inbox:               false,
				ReplyTo:             0,
				MessageAction:       0,
				MessageActionData:   nil,
				Entities:            nil,
				MediaType:           0,
				Media:               nil,
			}, userID)
			c.So(err, ShouldBeNil)
		}

		err = repo.Messages.ClearHistory(userID, 0, peerID, 1, 995)
		c.So(err, ShouldBeNil)

		d, err := repo.Dialogs.Get(0, peerID, 1)
		c.So(err, ShouldBeNil)
		c.So(d.TopMessageID, ShouldEqual, 999)

		ums, users, groups := repo.Messages.GetMessageHistory(0, peerID, 1, 0, 0, 100)
		c.So(users, ShouldNotBeNil)
		c.So(groups, ShouldNotBeNil)
		c.So(ums, ShouldHaveLength, 4)

		repo.Messages.Delete(userID, 0, peerID, 1, 999)
		d, err = repo.Dialogs.Get(0, peerID, 1)
		c.So(err, ShouldBeNil)
		c.So(d.TopMessageID, ShouldEqual, 998)
	})

}

func TestSearch(t *testing.T) {
	m := make([]*msg.UserMessage, 0, 10)
	for i := 1; i < 100; i++ {
		peerID := int64(i%10 + 1)
		peerType := int32(msg.PeerType_PeerUser)
		if i%2 == 0 {
			peerID = -peerID
			peerType = int32(msg.PeerType_PeerGroup)
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

	_ = repo.Messages.SearchTextByPeerID(0, "H", 6, 100)
	_ = repo.Messages.SearchTextByPeerID(0, "H", -7, 100)
}

func TestUserPhotoGallery(t *testing.T) {
	Convey("UserPhotoGallery", t, func(c C) {
		userID := tools.RandomInt64(0)
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
			ID:           userID,
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

		_, err := repo.Users.Get(userID)
		c.So(err, ShouldBeNil)

		user.PhotoGallery = []*msg.UserPhoto{photo1, photo2}
		repo.Users.Save(user)
		_, err = repo.Users.Get(userID)
		c.So(err, ShouldBeNil)

		phGallery := repo.Users.GetPhotoGallery(userID)
		c.So(phGallery, ShouldHaveLength, 2)
		c.So(phGallery[0].PhotoID, ShouldEqual, 1)
		c.So(phGallery[1].PhotoID, ShouldEqual, 2)
	})
}

func TestGroupPhotoGallery(t *testing.T) {
	Convey("GroupPhotoGallery", t, func(c C) {
		groupID := -tools.RandomInt64(0)
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
				ID:           groupID,
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

		phGallery, err := repo.Groups.GetPhotoGallery(groupID)
		c.So(err, ShouldBeNil)
		c.So(phGallery, ShouldHaveLength, 2)
	})

}

func TestMessagesSave(t *testing.T) {
	Convey("MessageSave", t, func(c C) {
		m := createMediaMessage("Hello", "file.txt", nil)
		repo.Messages.Save(m)
		media := &msg.MediaDocument{}
		err := media.Unmarshal(m.Media)
		c.So(err, ShouldBeNil)
		clientFile, err := repo.Files.Get(media.Doc.ClusterID, media.Doc.ID, media.Doc.AccessHash)
		c.So(err, ShouldBeNil)
		c.So(clientFile.MessageID, ShouldEqual, m.ID)
	})

}
