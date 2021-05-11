package gif

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/module"
)

/*
   Creation Time: 2021 - May - 10
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

type gif struct {
	module.Base
}

func New() *gif {
	r := &gif{}
	r.RegisterHandlers(
		map[int64]domain.LocalHandler{
			msg.C_GifDelete:   r.gifDelete,
			msg.C_GifGetSaved: r.gifGetSaved,
			msg.C_GifSave:     r.gifSave,
		},
	)
	r.RegisterMessageAppliers(
		map[int64]domain.MessageApplier{
			msg.C_SavedGifs: r.savedGifs,
		},
	)
	return r
}

func (r *gif) Name() string {
	return module.Gif
}
