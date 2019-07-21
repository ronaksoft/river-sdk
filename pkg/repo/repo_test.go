package repo_test

import (
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"testing"
	"time"
)

/*
   Creation Time: 2019 - Jul - 20
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
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
	// err := repo.Dialogs.Save(dialog)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	repo.Dialogs.SaveNew(dialog, time.Now().Unix())

	d := repo.Dialogs.Get(100, 1)
	t.Log(dialog)
	t.Log(d)
}

func TestRepoMessagesExtra_SaveScrollID(t *testing.T) {
	repo.MessagesExtra.SaveScrollID(11, 1, 101)
	x := repo.MessagesExtra.GetScrollID(11,1)
	fmt.Println(x)
}
