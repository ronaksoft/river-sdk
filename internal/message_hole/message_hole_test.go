package messageHole

import (
	"git.ronaksoft.com/river/sdk/internal/repo"
	"github.com/ronaksoft/rony/tools"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func init() {
	repo.InitRepo("./_data", false)
}
func TestHole(t *testing.T) {
	Convey("Hole", t, func(c C) {
		peerID := tools.RandomInt64(0)
		peerType := int32(1)

		Convey("Test 1", func(c C) {
			// Test 1
			peerID = tools.RandomInt64(0)
			InsertFill(0, peerID, peerType, 0, 10, 11)
			InsertFill(0, peerID, peerType, 0, 11, 13)
			InsertFill(0, peerID, peerType, 0, 15, 16)
			InsertFill(0, peerID, peerType, 0, 17, 19)
			c.So(IsHole(0, peerID, peerType, 0, 10, 14), ShouldBeFalse)
			fill, r := GetLowerFilled(0, peerID, peerType, 0, 16)
			c.So(fill, ShouldBeTrue)
			c.So(r.Min, ShouldEqual, 15)
			c.So(r.Max, ShouldEqual, 16)
			// _ ,_ = c.Println(PrintHole(0, peerID, peerType))
		})

		Convey("Test 2", func(c C) {
			peerID = tools.RandomInt64(0)
			InsertFill(0, peerID, peerType, 0, 6, 8)
			InsertFill(0, peerID, peerType, 0, 19, 20)
			InsertFill(0, peerID, peerType, 0, 12, 12)
			InsertFill(0, peerID, peerType, 0, 12, 12)
			InsertFill(0, peerID, peerType, 0, 15, 15)
			InsertFill(0, peerID, peerType, 0, 13, 14)
			// _, _ = c.Println(PrintHole(0, peerID, peerType))
			fill, _ := GetLowerFilled(0, peerID, peerType, 0, 21)
			c.So(fill, ShouldBeFalse)
			fill, r := GetUpperFilled(0, peerID, peerType, 0, 12)
			c.So(fill, ShouldBeTrue)
			c.So(r.Min, ShouldEqual, 12)
			c.So(r.Max, ShouldEqual, 15)

		})

		Convey("Test 3", func(c C) {
			// Test 3
			peerID = tools.RandomInt64(0)
			InsertFill(0, peerID, peerType, 0, 12, 12)
			InsertFill(0, peerID, peerType, 0, 101, 120)
			InsertFill(0, peerID, peerType, 0, 110, 120)
			InsertFill(0, peerID, peerType, 0, 140, 141)
			InsertFill(0, peerID, peerType, 0, 141, 142)
			InsertFill(0, peerID, peerType, 0, 143, 143)
			// _, _ = c.Println(PrintHole(0, peerID, peerType))
			fill, r := GetLowerFilled(0, peerID, peerType, 0, 141)
			c.So(fill, ShouldBeTrue)
			c.So(r.Max, ShouldEqual, 141)
			c.So(r.Min, ShouldEqual, 140)
			fill, r = GetUpperFilled(0, peerID, peerType, 0, 120)
			c.So(fill, ShouldBeTrue)
			c.So(r.Max, ShouldEqual, 120)
			c.So(r.Min, ShouldEqual, 120)

		})

		Convey("Test 4", func(c C) {
			peerID = tools.RandomInt64(0)
			InsertFill(0, peerID, peerType, 0, 1001, 1001)
			InsertFill(0, peerID, peerType, 0, 800, 900)
			InsertFill(0, peerID, peerType, 0, 700, 850)
			InsertFill(0, peerID, peerType, 0, 700, 799)
			InsertFill(0, peerID, peerType, 0, 701, 799)
			InsertFill(0, peerID, peerType, 0, 701, 801)
			InsertFill(0, peerID, peerType, 0, 100, 699)
			// _, _ = c.Println(PrintHole(0, peerID, peerType))
			fill, r := GetUpperFilled(0, peerID, peerType, 0, 700)
			c.So(fill, ShouldBeTrue)
			c.So(r.Min, ShouldEqual, 700)
			c.So(r.Max, ShouldEqual, 900)

			fill, r = GetUpperFilled(0, peerID, peerType, 0, 699)
			c.So(fill, ShouldBeTrue)
			c.So(r.Min, ShouldEqual, 699)
			c.So(r.Max, ShouldEqual, 900)
		})

		Convey("Test 5", func(c C) {
			peerID = tools.RandomInt64(0)
			InsertFill(0, peerID, peerType, 0, 1001, 1001)
			InsertFill(0, peerID, peerType, 0, 400, 500)
			InsertFill(0, peerID, peerType, 0, 600, 700)
			InsertFill(0, peerID, peerType, 0, 399, 699)
			// _, _ = c.Println(PrintHole(0, peerID, peerType))

			fill, r := GetUpperFilled(0, peerID, peerType, 0, 699)
			c.So(fill, ShouldBeTrue)
			c.So(r.Min, ShouldEqual, 699)
			c.So(r.Max, ShouldEqual, 700)
		})

	})

}
