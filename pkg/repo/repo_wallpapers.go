package repo

import (
	msg "git.ronaksoftware.com/river/msg/chat"
	"github.com/dgraph-io/badger"
)

type repoWallpapers struct {
	*repository
}

func (r *repoWallpapers) SaveWallpapers(wallpapers *msg.WallPapersMany) error {
	if len(wallpapers.WallPapers) == 0 {
		return nil
	}

	err := badgerUpdate(func(txn *badger.Txn) error {
		for _, o := range wallpapers.WallPapers {
			err := Files.SaveWallpaper(txn, o)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}
