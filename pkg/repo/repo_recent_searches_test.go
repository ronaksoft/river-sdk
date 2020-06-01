package repo_test

import (
	msg "git.ronaksoftware.com/river/msg/chat"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

/**
 * @created 01/06/2020 - 14:55
 * @project riversdk
 * @author reza
 */

func TestRepoRecentSearches(t *testing.T) {
	Convey("Messages Search", t, func() {
		Convey("Search By Label", func(c C) {
			err := repo.RecentSearches.Put(&msg.RecentSearch{
				Peer: &msg.Peer{
					ID:         101,
					Type:       1,
					AccessHash: 1010,
				},
				Date: int32(time.Now().Unix()),
			})
			c.So(err, ShouldBeNil)
			searches := repo.RecentSearches.List(1)
			c.So(searches, ShouldHaveLength, 1)
		})
	})
}
