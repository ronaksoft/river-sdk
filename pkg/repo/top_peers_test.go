package repo_test

import (
	"git.ronaksoft.com/river/msg/msg"
	"git.ronaksoft.com/river/sdk/internal/tools"
	"git.ronaksoft.com/river/sdk/pkg/repo"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

/*
   Creation Time: 2020 - Sep - 22
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func TestTopPeer(t *testing.T) {
	Convey("Testing TopPeers", t, func(c C) {
		Convey("Save TopPeer", func(c C) {
			for i := 0; i < 10; i++ {

				userID := tools.RandomInt64(0)
				teamID := int64(0)
				var topPeers []*msg.TopPeer
				for i := 0; i < 10; i++ {
					topPeers = append(topPeers, &msg.TopPeer{
						TeamID: teamID,
						Peer: &msg.Peer{
							ID:         tools.RandomInt64(0),
							Type:       1,
							AccessHash: 0,
						},
						Rate:       1.0 / float32(i),
						LastUpdate: time.Now().Unix(),
					})
				}
				err := repo.TopPeers.Save(msg.TopPeerCategory_Forwards, userID, teamID, topPeers...)
				c.So(err, ShouldBeNil)

				list, err := repo.TopPeers.List(teamID, msg.TopPeerCategory_Forwards, 0, 10)
				c.So(err, ShouldBeNil)
				c.So(list, ShouldHaveLength, len(topPeers))
			}
		})

	})
}
