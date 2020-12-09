package repo_test

import (
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/tools"
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
			userID := tools.RandomInt64(0)
			teamID := int64(0)
			var topPeers []*msg.TopPeer
			for i := 0; i < 10; i++ {
				peerID := tools.RandomInt64(0)
				topPeers = append(topPeers, &msg.TopPeer{
					TeamID: teamID,
					Peer: &msg.Peer{
						ID:         peerID,
						Type:       1,
						AccessHash: 0,
					},
					Rate:       1.0 / float32(i),
					LastUpdate: time.Now().Unix(),
				})
				_ = repo.Users.Save(&msg.User{
					ID:        peerID,
					FirstName: fmt.Sprintf("User%d", i),
				})
			}

			for i := 0; i < 10; i++ {
				for _, tp := range topPeers {
					err := repo.TopPeers.Update(msg.TopPeerCategory_Forwards, userID, teamID, tp.Peer.ID, tp.Peer.Type)
					c.So(err, ShouldBeNil)
					err = repo.TopPeers.Update(msg.TopPeerCategory_Users, userID, teamID, tp.Peer.ID, tp.Peer.Type)
					c.So(err, ShouldBeNil)
				}
				time.Sleep(time.Second)
				for _, tp := range topPeers {
					err := repo.TopPeers.Update(msg.TopPeerCategory_Forwards, userID, teamID, tp.Peer.ID, tp.Peer.Type)
					c.So(err, ShouldBeNil)
					err = repo.TopPeers.Update(msg.TopPeerCategory_Users, userID, teamID, tp.Peer.ID, tp.Peer.Type)
					c.So(err, ShouldBeNil)
				}

				list, err := repo.TopPeers.List(teamID, msg.TopPeerCategory_Forwards, 0, 10)
				c.So(err, ShouldBeNil)
				c.So(list, ShouldHaveLength, len(topPeers))

			}
		})

	})
}
