package repo_test

import (
	"git.ronaksoft.com/river/msg/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/repo"
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
				doc := &msg.Document{
					ID:         docID,
					ClusterID:  clusterID,
					AccessHash: 100,
					FileSize:   0,
				}
				md := &msg.MediaDocument{
					Doc:          doc,
					Caption:      "",
					TTLinSeconds: 0,
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
