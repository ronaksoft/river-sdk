package repo_test

import (
	"git.ronaksoftware.com/river/msg/msg"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

/*
   Creation Time: 2020 - Jun - 15
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func TestGif(t *testing.T) {
	Convey("Testing GIFs", t, func(c C) {
		Convey("Save GIF", func(c C) {
			for i := 0; i < 10; i++ {
				clusterID := int32(domain.RandomInt64(100))
				docID := domain.RandomInt64(0)
				md := &msg.MediaDocument{
					Caption:      "Some Caption",
					TTLinSeconds: 0,
					Doc:          &msg.Document{
						ID:          docID,
						AccessHash:  domain.RandomUint64(),
						Date:        0,
						MimeType:    "",
						FileSize:    0,
						Version:     0,
						ClusterID:   clusterID,
						Attributes:  nil,
						Thumbnail:   nil,
						MD5Checksum: "",
					},
					Entities:     nil,
				}
				err := repo.Gifs.Save(md)
				c.So(err, ShouldBeNil)
				err = repo.Gifs.UpdateLastAccess(clusterID, docID, domain.Now().Unix())
				c.So(err, ShouldBeNil)
				found := repo.Gifs.IsSaved(clusterID, docID)
				c.So(found, ShouldBeTrue)
			}

			savedGifs, err := repo.Gifs.GetSaved()
			c.So(err, ShouldBeNil)
			c.So(len(savedGifs.Docs), ShouldBeGreaterThan, 0)
		})

	})
}
