package repo_test

import (
	"git.ronaksoft.com/ronak/riversdk/pkg/domain"
	"git.ronaksoft.com/ronak/riversdk/pkg/repo"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

/*
   Creation Time: 2019 - Dec - 28
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func TestMessagesSearch(t *testing.T) {
	for i := 1; i < 100; i++ {
		repo.Messages.Save(createMessage(int64(i), domain.RandomID(32), []int32{int32(i % 5)}))
	}
	Convey("Messages Search", t, func() {
		Convey("Search By Label", func(c C) {
			msgs := repo.Messages.SearchByLabels([]int32{1}, 0, 100)
			for _, m := range msgs {
				_, _ = c.Println(m.ID, m.Body, m.LabelIDs)
			}
		})
	})
}
