package user

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

type user struct {
	module.Base
}

func New() *user {
	r := &user{}
	r.RegisterHandlers(
		map[int64]domain.LocalMessageHandler{
			msg.C_UsersGet:     r.usersGet,
			msg.C_UsersGetFull: r.usersGetFull,
		},
	)
	r.RegisterUpdateAppliers(
		map[int64]domain.UpdateApplier{
			msg.C_UpdateUsername:    r.updateUsername,
			msg.C_UpdateUserBlocked: r.updateUserBlocked,
			msg.C_UpdateUserPhoto:   r.updateUserPhoto,
		},
	)
	r.RegisterMessageAppliers(
		map[int64]domain.MessageApplier{
			msg.C_UsersMany: r.usersMany,
		},
	)
	return r
}

func (r *user) Name() string {
	return module.User
}
