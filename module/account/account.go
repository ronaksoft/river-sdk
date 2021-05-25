package account

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

type account struct {
	module.Base
}

func New() *account {
	r := &account{}
	r.RegisterHandlers(
		map[int64]request.LocalHandler{
			msg.C_AccountGetTeams:          r.accountsGetTeams,
			msg.C_AccountRegisterDevice:    r.accountRegisterDevice,
			msg.C_AccountRemovePhoto:       r.accountRemovePhoto,
			msg.C_AccountSetNotifySettings: r.accountSetNotifySettings,
			msg.C_AccountUnregisterDevice:  r.accountUnregisterDevice,
			msg.C_AccountUpdateProfile:     r.accountUpdateProfile,
			msg.C_AccountUpdateUsername:    r.accountUpdateUsername,
		},
	)
	r.RegisterUpdateAppliers(map[int64]domain.UpdateApplier{
		msg.C_UpdateAccountPrivacy: r.updateAccountPrivacy,
	})
	return r
}

func (r *account) Name() string {
	return module.Account
}
