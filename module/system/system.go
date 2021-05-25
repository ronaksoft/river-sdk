package system

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/request"
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

type system struct {
	module.Base
}

func New() *system {
	r := &system{}
	r.RegisterHandlers(
		map[int64]request.LocalHandler{
			msg.C_SystemGetConfig: r.systemGetConfig,
		},
	)
	r.RegisterMessageAppliers(
		map[int64]domain.MessageApplier{
			msg.C_SystemConfig: r.systemConfig,
		},
	)
	return r
}

func (r *system) Name() string {
	return module.System
}
