package repo_test

import (
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"
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
	err := repo.InitRepo("./_data")
	if err != nil {
		logs.Fatal(err.Error())
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

	d := repo.Dialogs.Get(100, 1)
	t.Log(dialog)
	t.Log(d)
}

func TestRepoMessagesExtra(t *testing.T) {
	repo.MessagesExtra.SaveScrollID(11, 1, 101)
	x := repo.MessagesExtra.GetScrollID(11, 1)
	fmt.Println(x)
}

func TestRepoFiles(t *testing.T) {
	fs := &dto.FilesStatus{
		MessageID:     1,
		FileID:        10,
		AccessHash:    11,
		ClusterID:     2,
		RequestStatus: 1,
	}
	repo.Files.SaveStatus(fs)

	fs2, err := repo.Files.GetStatus(1)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(fs2.RequestStatus)

	fs.RequestStatus = 2
	repo.Files.SaveStatus(fs)

	fs2, err = repo.Files.GetStatus(1)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(fs2.RequestStatus)

	s := repo.Files.GetAllStatuses()
	fmt.Println(s)

}

func TestPending(t *testing.T) {
	pm := new(msg.ClientSendMessageMedia)
	pm.Peer = new(msg.InputPeer)
	_, err := repo.PendingMessages.SaveClientMessageMedia(10, 1, 11, pm)
	if err != nil {
		t.Error(err)
	}
	pm1 := repo.PendingMessages.GetByID(10)
	fmt.Println(pm1)

	repo.PendingMessages.Delete(10)

	pm2 := repo.PendingMessages.GetByID(10)
	fmt.Println(pm2)
}
