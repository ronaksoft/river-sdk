package repo

import (
	"fmt"
	"log"
	"testing"

	"git.ronaksoftware.com/ronak/riversdk/msg"
)

func TestUsersSaveMany(t *testing.T) {
	// Create DB and Tables
	err := InitRepo("sqlite3", "river.db")
	if err != nil {
		log.Fatal(err)
	}
	Ctx().LogMode(true)
	defer Ctx().Close()
	//defer os.Remove("river.db")

	users := make([]*msg.User, 0)
	for i := 1; i < 4; i++ {
		users = append(users, &msg.User{
			ID:        int64(i),
			FirstName: fmt.Sprintf("FirstName[%d]AAA", i),
			LastName:  fmt.Sprintf("LastName[%d]", i),
			Username:  fmt.Sprintf("UserName[%d]", i),
		})
	}

	err = Ctx().Users.SaveMany(users)
	if err != nil {
		t.Error(err)
	}
	testID := []int64{1, 2}
	fetch := Ctx().Users.GetAnyUsers(testID)
	if fetch == nil {
		t.Failed()
	}
	if len(fetch) != 2 {
		t.Failed()
	}
}
