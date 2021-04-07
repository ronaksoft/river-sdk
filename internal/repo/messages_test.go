package repo_test

import (
	"git.ronaksoft.com/river/msg/go/msg"
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

func TestGetMediaMessageHistory(t *testing.T) {
	Convey("GetMediaHistory", t, func(c C) {
		teamID := tools.RandomInt64(0)
		peerID := tools.RandomInt64(0)
		userID := tools.RandomInt64(0)
		err := repo.Dialogs.SaveNew(&msg.Dialog{
			TeamID:       teamID,
			PeerID:       peerID,
			PeerType:     1,
			TopMessageID: 0,
		}, tools.TimeUnix())
		c.So(err, ShouldBeNil)

		for i := int64(1); i <= 100; i++ {
			err = repo.Messages.SaveNew(&msg.UserMessage{
				TeamID:   teamID,
				PeerID:   peerID,
				PeerType: 1,
				ID:       i,
				MediaCat: msg.MediaCategory_MediaCategoryAudio,
				SenderID: peerID,
			}, userID)
			c.So(err, ShouldBeNil)
		}

		Convey("Load With MaxID = 0 and MinID = 0", func(c C) {
			ums, _, _ := repo.Messages.GetMediaMessageHistory(teamID, peerID, 1, 0, 0, 10, msg.MediaCategory_MediaCategoryAudio)
			c.So(ums, ShouldHaveLength, 10)
			c.So(ums[0].ID, ShouldEqual, 100)
			c.So(ums[9].ID, ShouldEqual, 91)
		})
		Convey("Load With MaxID > 0", func(c C) {
			ums, _, _ := repo.Messages.GetMediaMessageHistory(teamID, peerID, 1, 0, 50, 10, msg.MediaCategory_MediaCategoryAudio)
			c.So(ums, ShouldHaveLength, 10)
			c.So(ums[0].ID, ShouldEqual, 50)
			c.So(ums[9].ID, ShouldEqual, 41)
		})
		Convey("Load With MinID > 0", func(c C) {
			ums, _, _ := repo.Messages.GetMediaMessageHistory(teamID, peerID, 1, 30, 0, 10, msg.MediaCategory_MediaCategoryAudio)
			c.So(ums, ShouldHaveLength, 10)
			c.So(ums[0].ID, ShouldEqual, 30)
			c.So(ums[9].ID, ShouldEqual, 39)
		})
	})
}
