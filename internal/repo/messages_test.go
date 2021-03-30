package repo_test

import (
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"github.com/ronaksoft/rony/tools"
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
	peerID := tools.RandomInt64(0)
	for i := 1; i < 100; i++ {
		repo.Messages.Save(createMessage(int64(i), peerID, domain.RandomID(32), []int32{int32(i % 5)}))
	}
	Convey("Messages Search", t, func() {
		Convey("Search By Label", func(c C) {
			msgs := repo.Messages.SearchByLabels(0, []int32{1}, peerID, 100)
			c.So(msgs, ShouldHaveLength, 20)
		})
	})
}
