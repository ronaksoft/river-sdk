package repo

import (
	"fmt"
	"log"
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
	fetch, err := Ctx().Groups.GetManyGroups(testID)
	if err != nil {
		t.Error(err)
	}
	if len(fetch) != 2 {
		t.Failed()
	}
}
