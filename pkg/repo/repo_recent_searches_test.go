package repo_test

import (
	"git.ronaksoftware.com/river/msg/msg"
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
	Convey("RecentSearch Repo", t, func() {
		Convey("Put, List and Clear", func(c C) {
			err := repo.RecentSearches.Clear()
			c.So(err, ShouldBeNil)
			err = repo.RecentSearches.Put(&msg.RecentSearch{
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
			err = repo.RecentSearches.Put(&msg.RecentSearch{
				Peer: &msg.Peer{
					ID:         102,
					Type:       1,
					AccessHash: 1020,
				},
				Date: int32(time.Now().Unix()),
			})
			c.So(err, ShouldBeNil)
			searches = repo.RecentSearches.List(2)
			c.So(searches, ShouldHaveLength, 2)
			err = repo.RecentSearches.Clear()
			c.So(err, ShouldBeNil)
			searches = repo.RecentSearches.List(2)
			c.So(searches, ShouldHaveLength, 0)
		})
	})
}
