package user

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/module"
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
	return r
}
