package minirepo_test

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/minirepo"
	"github.com/ronaksoft/rony/tools"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

/*
   Creation Time: 2021 - May - 05
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func TestRepoDialogs(t *testing.T) {
	Convey("MiniRepo Dialogs", t, func(c C) {
		for i := 0; i < 10; i++ {
			err := minirepo.Dialogs.Save(&msg.Dialog{
				TeamID:       0,
				PeerID:       tools.RandomInt64(0),
				PeerType:     int32(tools.RandomInt(2)) + 1,
				TopMessageID: tools.RandomInt64(100000),
			})
			c.So(err, ShouldBeNil)
		}
		dialogs, err := minirepo.Dialogs.List(0, 0, 10)
		c.So(err, ShouldBeNil)
		c.So(dialogs, ShouldHaveLength, 10)
	})

}
