package repo

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/msg"
)

func TestGroupsSaveMany(t *testing.T) {
	// Create DB and Tables
	err := InitRepo("sqlite3", "river.db")
	if err != nil {
		log.Fatal(err)
	}
	Ctx().LogMode(true)
	defer Ctx().Close()
	//defer os.Remove("river.db")

	groups := make([]*msg.Group, 0)
	for i := 1; i < 4; i++ {
		groups = append(groups, &msg.Group{
			CreatedOn: time.Now().Unix(),
			ID:        int64(i),
			Title:     fmt.Sprintf("Title AAA [%d]", i),
		})
	}

	err = Ctx().Groups.SaveMany(groups)
	if err != nil {
		t.Error(err)
	}
	testID := []int64{1, 2}
	fetch := Ctx().Groups.GetManyGroups(testID)
	if len(fetch) != 2 {
		t.Failed()
	}
}

func TestAddGroupMember(t *testing.T) {
	// Create DB and Tables
	err := InitRepo("sqlite3", "river.db")
	if err != nil {
		log.Fatal(err)
	}
	Ctx().LogMode(true)
	defer Ctx().Close()
	//defer os.Remove("river.db")

	for i := 1; i < 11; i++ {
		x := new(msg.UpdateGroupMemberAdded)
		x.GroupID = int64(i)
		x.ChatVersion = int32(i)
		x.Date = time.Now().Unix()
		x.InviterID = int64(i)
		x.UserID = int64(i)
		err = Ctx().Groups.AddGroupMember(x)
		if err != nil {
			t.Fail()
		}
	}

	gp, err := Ctx().Groups.GetParticipants(1)
	if err != nil {
		t.Fail()
	}
	fmt.Println(gp)
}

func TestDeleteGroupMember(t *testing.T) {
	// Create DB and Tables
	os.Remove("river.db")
	err := InitRepo("sqlite3", "river.db")
	if err != nil {
		log.Fatal(err)
	}
	Ctx().LogMode(true)
	defer Ctx().Close()
	//defer os.Remove("river.db")

	x := new(msg.UpdateGroupMemberAdded)
	x.GroupID = 1
	x.ChatVersion = 2
	x.Date = time.Now().Unix()
	x.InviterID = 3
	x.UserID = 4

	err = Ctx().Groups.AddGroupMember(x)
	if err != nil {
		t.Fail()
	}

	err = Ctx().Groups.DeleteGroupMember(x.ChatID, x.UserID)
	if err != nil {
		t.Fail()
	}

	gp, err := Ctx().Groups.GetParticipants(x.ChatID)
	if err != nil {
		t.Fail()
	}
	if len(gp) != 0 {
		t.Fail()
	}
}

func TestUpdateGroupTitle(t *testing.T) {
	// Create DB and Tables
	err := InitRepo("sqlite3", "river.db")
	if err != nil {
		log.Fatal(err)
	}
	Ctx().LogMode(true)
	defer Ctx().Close()
	//defer os.Remove("river.db")

	x := new(msg.Group)
	x.CreatedOn = time.Now().Unix()
	x.EditedOn = 0
	x.ID = 113
	x.Participants = 10
	x.Title = "OIOIOIOIOI"

	err = Ctx().Groups.Save(x)
	if err != nil {
		t.Fail()
	}
	err = Ctx().Groups.UpdateGroupTitle(113, "ZZZZZZZZ")
	if err != nil {
		t.Fail()
	}
	gs := Ctx().Groups.GetManyGroups([]int64{113})
	if gs[0].Title != "ZZZZZZZZ" {

		t.Fail()
	}
}

func TestSaveParticipants(t *testing.T) {
	// Create DB and Tables
	err := InitRepo("sqlite3", "river.db")
	if err != nil {
		log.Fatal(err)
	}
	Ctx().LogMode(true)
	defer Ctx().Close()
	//defer os.Remove("river.db")

	x := new(msg.GroupParticipant)
	x.Date = time.Now().Unix()
	x.InviterID = 113
	x.Type = msg.ParticipantType_Admin
	x.UserID = 117

	err = Ctx().Groups.SaveParticipants(113113, x)
	if err != nil {
		t.Fail()
	}
	err = Ctx().Groups.SaveParticipants(113113, x)
	if err != nil {
		t.Fail()
	}

	ps, err := Ctx().Groups.GetParticipants(113113)
	if err != nil {
		t.Fail()
	}
	if len(ps) != 1 {
		t.Fail()
	}
}
