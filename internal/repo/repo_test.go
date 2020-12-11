package repo_test

import (
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/repo"
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
	repo.MustInitRepo("./_data", false)
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
func createMessage(id int64, body string, labelIDs []int32) *msg.UserMessage {
	userID := domain.RandomInt63()

	return &msg.UserMessage{
		ID:                  id,
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
	repo.MessagesExtra.SaveScrollID(0, 11, 1, 101)
	x := repo.MessagesExtra.GetScrollID(0, 11, 1)
	fmt.Println(x)
}

func TestPending(t *testing.T) {
	pm := new(msg.ClientSendMessageMedia)
	pm.Peer = new(msg.InputPeer)
	_, err := repo.PendingMessages.SaveClientMessageMedia(0, 0, 10, 1, 11, 20, 21, pm, nil)
	if err != nil {
		t.Error(err)
	}
	pm1 := repo.PendingMessages.GetByID(10)
	fmt.Println(pm1)

	_ = repo.PendingMessages.Delete(10)
	pm2 := repo.PendingMessages.GetByID(10)
	fmt.Println(pm2)
}

func TestGetMessageKey(t *testing.T) {
	peerID := 10001
	peerType := 1
	msgID := 1 << 32
	domain.StrToByte(fmt.Sprintf("%s.%021d.%d.%012d", "MSG", peerID, peerType, msgID))
	fmt.Println(fmt.Sprintf("%s.%021d.%d.%012d", "MSG", peerID, peerType, msgID))
}

func TestRepoDeleteMessage(t *testing.T) {
	peerID := int64(10001)
	peerType := int32(1)

	d, _ := repo.Dialogs.Get(0, peerID, peerType)
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

	d, _ = repo.Dialogs.Get(0, peerID, peerType)
	fmt.Println(d)

	repo.Messages.Delete(10002, 0, peerID, peerType, 19)
	d, _ = repo.Dialogs.Get(0, peerID, peerType)
	fmt.Println(d)

	msgs, _ := repo.Messages.GetMessageHistory(0, peerID, peerType, 0, 0, 5)
	for idx := range msgs {
		fmt.Println(msgs[idx].ID)
	}

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
				logs.Fatal("Error On Save Pending", zap.Error(err))
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
	err = repo.Messages.ClearHistory(101, 0, 10, 1, 995)
	if err != nil {
		t.Error(err)
		return
	}
	d, err := repo.Dialogs.Get(0, 10, 1)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(d.TopMessageID)
	ums, us := repo.Messages.GetMessageHistory(0, 10, 1, 0, 0, 100)
	fmt.Println(len(ums), len(us))

	var x []int64
	for _, um := range ums {
		x = append(x, um.ID)
	}
	fmt.Println(x)

	repo.Messages.Delete(101, 10, 1, 1950)
	d, err = repo.Dialogs.Get(0, 10, 1)
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
	fmt.Println("Saved")

	// mm := repo.Messages.SearchText("Hello")
	fmt.Print("Search in UserPeer:")
	mm := repo.Messages.SearchTextByPeerID(0, "H", 6, 100)
	for _, m := range mm {
		fmt.Println(m.ID, m.Body, m.PeerID)
	}
	fmt.Print("Search in GroupPeer:")
	mm = repo.Messages.SearchTextByPeerID(0, "H", -7, 100)
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

	phGallery, err := repo.Groups.GetPhotoGallery(1000)
	if err != nil {
		t.Fatal(err)
	}
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
	t.Log(repo.Files.GetFilePath(clientFile))
}
