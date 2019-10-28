package repo_test

import (
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	ronak "git.ronaksoftware.com/ronak/toolbox"
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

func TestGetUserMessageKey(t *testing.T) {
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
		repo.Messages.SaveNew(m, d, 10002)
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
	for i := 1; i < 1000 ;i++ {
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
	// m := make([]*msg.UserMessage, 0, 10)
	// for i := 1; i < 1000 ;i++ {
	// 	m = append(m, &msg.UserMessage{
	// 		ID:                  int64(i),
	// 		PeerID:              14,
	// 		PeerType:            1,
	// 		CreatedOn:           time.Now().Unix(),
	// 		EditedOn:            0,
	// 		FwdSenderID:         0,
	// 		FwdChannelID:        0,
	// 		FwdChannelMessageID: 0,
	// 		Flags:               0,
	// 		MessageType:         0,
	// 		Body:                fmt.Sprintf("Hello %d", i),
	// 		SenderID:            100,
	// 		ContentRead:         false,
	// 		Inbox:               false,
	// 		ReplyTo:             0,
	// 		MessageAction:       0,
	// 		MessageActionData:   nil,
	// 		Entities:            nil,
	// 		MediaType:           0,
	// 		Media:               nil,
	// 	})
	// }
	// repo.Messages.Save(m...)
	// fmt.Println("Saved")


	// mm := repo.Messages.SearchText("Hello")
	mm := repo.Messages.SearchTextByPeerID("H", 6)
	for _, m := range mm {
		fmt.Println(m.ID, m.Body, m.PeerID)
	}
}