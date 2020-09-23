package fileCtrl

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

/*
   Creation Time: 2020 - Sep - 23
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func TestBestChunkSize(t *testing.T) {
	Convey("BestChunkSize", t, func(c C) {
		c.Println("500KB", bestChunkSize(500*1024))
		c.Println("5MB", bestChunkSize(5*1024*1024))
		c.Println("30MB", bestChunkSize(30*1024*1024))
		c.Println("100MB", bestChunkSize(100*1024*1024))
	})
}
