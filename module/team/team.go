package team

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

type team struct {
	module.Base
}

func New() *team {
	r := &team{}
	r.RegisterHandlers(
		map[int64]domain.LocalMessageHandler{
			msg.C_TeamEdit:              r.teamEdit,
			msg.C_ClientGetTeamCounters: r.clientGetTeamCounters,
		},
	)
	return r
}
