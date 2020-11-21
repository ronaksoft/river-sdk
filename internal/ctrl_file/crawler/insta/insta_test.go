package insta

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

/*
   Creation Time: 2020 - Nov - 21
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func TestGetPostInfo(t *testing.T) {
	Convey("GetPostInfo", t, func(c C) {
		em, err := GetPostInfo("https://www.instagram.com/p/CEJehvrA69J/")
		c.So(err, ShouldBeNil)
		c.Println(em.Typename, em.DisplayUrl)
	})

}
