package repo_test

import (
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"github.com/ronaksoft/rony/tools"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

/*
   Creation Time: 2020 - Jun - 03
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func TestLabel(t *testing.T) {
	Convey("Testing Labels", t, func(c C) {
		Convey("Save Labels", func(c C) {
			err := repo.Labels.Save(
				0,
				&msg.Label{
					ID:     1,
					Name:   "Label1",
					Colour: "#FF0000",
				},
				&msg.Label{
					ID:     2,
					Name:   "Label 2",
					Colour: "#00FF00",
				},
			)
			c.So(err, ShouldBeNil)
			labels, err := repo.Labels.GetAll(0)
			c.So(err, ShouldBeNil)
			c.So(labels, ShouldHaveLength, 2)
		})
		Convey("Add Label To Message", func(c C) {
			peerID := tools.RandomInt64(0)
			// Save 10 new message for peerID
			for i := 1; i <= 10; i++ {
				repo.Messages.Save(&msg.UserMessage{
					ID:                  int64(i),
					PeerID:              peerID,
					PeerType:            0,
					CreatedOn:           0,
					EditedOn:            0,
					FwdSenderID:         0,
					FwdChannelID:        0,
					FwdChannelMessageID: 0,
					Flags:               0,
					MessageType:         0,
					Body:                fmt.Sprintf("Test %d", i),
					SenderID:            0,
					ContentRead:         false,
					Inbox:               false,
					ReplyTo:             0,
					MessageAction:       0,
					MessageActionData:   nil,
					Entities:            nil,
					MediaType:           0,
					Media:               nil,
					ReplyMarkup:         0,
					ReplyMarkupData:     nil,
					LabelIDs:            nil,
				})
			}

			ums, _, _ := repo.Labels.ListMessages(1, 0, 100, 0, 0)
			for _, um := range ums {
				err := repo.Labels.RemoveLabelsFromMessages([]int32{1}, 0, um.PeerID, um.PeerType, []int64{um.ID})
				c.So(err, ShouldBeNil)
			}

			err := repo.Labels.AddLabelsToMessages([]int32{1}, 0, peerID, 1, []int64{1, 2, 3, 6, 8, 9, 10})
			c.So(err, ShouldBeNil)
		})
		Convey("List Messages", func(c C) {
			ums, _, _ := repo.Labels.ListMessages(1, 0, 3, 0, 0)
			c.So(ums, ShouldHaveLength, 3)
			c.So(ums[0].ID, ShouldEqual, 1)
			c.So(ums[1].ID, ShouldEqual, 2)
			c.So(ums[2].ID, ShouldEqual, 3)

			ums, _, _ = repo.Labels.ListMessages(1, 0, 2, 6, 0)
			c.So(ums, ShouldHaveLength, 2)
			c.So(ums[0].ID, ShouldEqual, 6)
			c.So(ums[1].ID, ShouldEqual, 8)

			ums, _, _ = repo.Labels.ListMessages(1, 0, 3, 0, 9)
			c.So(ums, ShouldHaveLength, 3)
			c.So(ums[0].ID, ShouldEqual, 6)
			c.So(ums[1].ID, ShouldEqual, 8)
			c.So(ums[2].ID, ShouldEqual, 9)

			ums, _, _ = repo.Labels.ListMessages(1, 0, 3, 0, 5)
			c.So(ums, ShouldHaveLength, 3)
			c.So(ums[0].ID, ShouldEqual, 1)
			c.So(ums[1].ID, ShouldEqual, 2)
			c.So(ums[2].ID, ShouldEqual, 3)

			ums, _, _ = repo.Labels.ListMessages(1, 0, 3, 5, 0)
			c.So(ums, ShouldHaveLength, 3)
			c.So(ums[0].ID, ShouldEqual, 6)
			c.So(ums[1].ID, ShouldEqual, 8)
			c.So(ums[2].ID, ShouldEqual, 9)
		})
		Convey("Label Hole", func(c C) {
			b, _ := repo.Labels.GetLowerFilled(0, 10, 100)
			c.So(b, ShouldBeFalse)

			err := repo.Labels.Fill(0, 1, 10, 100)
			c.So(err, ShouldBeNil)

			bar := repo.Labels.GetFilled(0, 1)
			c.So(bar.MinID, ShouldEqual, 10)
			c.So(bar.MaxID, ShouldEqual, 100)

			b, bar = repo.Labels.GetUpperFilled(0, 1, 90)
			c.So(b, ShouldBeTrue)
			c.So(bar.MinID, ShouldEqual, 90)
			c.So(bar.MaxID, ShouldEqual, 100)

			b, bar = repo.Labels.GetLowerFilled(0, 1, 90)
			c.So(b, ShouldBeTrue)
			c.So(bar.MinID, ShouldEqual, 10)
			c.So(bar.MaxID, ShouldEqual, 90)
		})
		Convey("Search By Label", func(c C) {

		})
		Convey("List complete test", func(c C) {
			peerID := domain.RandomInt64(0)
			for i := 1; i <= 10; i++ {
				repo.Messages.Save(&msg.UserMessage{
					ID:                  int64(i),
					PeerID:              peerID,
					PeerType:            int32(msg.PeerType_PeerUser),
					CreatedOn:           0,
					EditedOn:            0,
					FwdSenderID:         0,
					FwdChannelID:        0,
					FwdChannelMessageID: 0,
					Flags:               0,
					MessageType:         0,
					Body:                fmt.Sprintf("Test %d", i),
					SenderID:            0,
					ContentRead:         false,
					Inbox:               false,
					ReplyTo:             0,
					MessageAction:       0,
					MessageActionData:   nil,
					Entities:            nil,
					MediaType:           0,
					Media:               nil,
					ReplyMarkup:         0,
					ReplyMarkupData:     nil,
					LabelIDs:            nil,
				})
			}
			ums, _, _ := repo.Labels.ListMessages(1, 0, 100, 0, 0)
			for _, um := range ums {
				err := repo.Labels.RemoveLabelsFromMessages([]int32{1}, 0, um.PeerID, um.PeerType, []int64{um.ID})
				c.So(err, ShouldBeNil)
			}
			err := repo.Labels.AddLabelsToMessages([]int32{1}, 0, peerID, int32(msg.PeerType_PeerUser), []int64{1, 2, 3})
			c.So(err, ShouldBeNil)
			ums, _, _ = repo.Labels.ListMessages(1, 0, 100, 0, 0)
			c.So(ums, ShouldHaveLength, 3)
			err = repo.Labels.AddLabelsToMessages([]int32{1}, 0, peerID, int32(msg.PeerType_PeerUser), []int64{8})
			c.So(err, ShouldBeNil)
			ums, _, _ = repo.Labels.ListMessages(1, 0, 100, 0, 0)
			c.So(ums, ShouldHaveLength, 4)
			err = repo.Labels.AddLabelsToMessages([]int32{1}, 0, peerID, int32(msg.PeerType_PeerUser), []int64{7})
			c.So(err, ShouldBeNil)
			ums, _, _ = repo.Labels.ListMessages(1, 0, 100, 0, 0)
			c.So(ums, ShouldHaveLength, 5)
		})
	})
}
