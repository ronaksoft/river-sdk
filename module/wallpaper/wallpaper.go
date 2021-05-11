package wallpaper

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/module"
	"github.com/ronaksoft/rony"
	"go.uber.org/zap"
)

/*
   Creation Time: 2021 - May - 10
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

type wallpaper struct {
	module.Base
}

func New() *wallpaper {
	r := &wallpaper{}
	r.RegisterMessageAppliers(
		map[int64]domain.MessageApplier{
			msg.C_WallPapersMany: r.wallpapersMany,
		},
	)
	return r
}

func (r *wallpaper) Name() string {
	return module.Wallpaper
}

func (r *wallpaper) wallpapersMany(e *rony.MessageEnvelope) {
	u := &msg.WallPapersMany{}
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal wallpapersMany", zap.Error(err))
		return
	}

	err = repo.Wallpapers.SaveWallpapers(u)
	if err != nil {
		logs.Error("SyncCtrl got error on saving wallpapersMany", zap.Error(err))
	}
}
