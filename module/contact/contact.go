package contact

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

type contact struct {
	module.Base
}

func New() *contact {
	r := &contact{}
	r.RegisterHandlers(
		map[int64]domain.LocalMessageHandler{
			msg.C_ContactsAdd:          r.contactsAdd,
			msg.C_ContactsDelete:       r.contactsDelete,
			msg.C_ContactsDeleteAll:    r.contactsDeleteAll,
			msg.C_ContactsGet:          r.contactsGet,
			msg.C_ContactsGetTopPeers:  r.contactsGetTopPeers,
			msg.C_ContactsImport:       r.contactsImport,
			msg.C_ContactsResetTopPeer: r.contactsResetTopPeer,
			msg.C_ClientContactSearch:  r.clientContactSearch,
		},
	)
	return r
}
