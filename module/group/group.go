package group

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

type group struct {
	module.Base
}

func New() *group {
	r := &group{}
	r.RegisterHandlers(
		map[int64]domain.LocalMessageHandler{
			msg.C_GroupsAddUser:      r.groupAddUser,
			msg.C_GroupsDeleteUser:   r.groupDeleteUser,
			msg.C_GroupsEditTitle:    r.groupsEditTitle,
			msg.C_GroupsGetFull:      r.groupsGetFull,
			msg.C_GroupsRemovePhoto:  r.groupRemovePhoto,
			msg.C_GroupsToggleAdmins: r.groupToggleAdmin,
			msg.C_GroupsUpdateAdmin:  r.groupUpdateAdmin,
		},
	)
	return r
}
