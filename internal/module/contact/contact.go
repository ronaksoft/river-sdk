package contact

import (
	"git.ronaksoft.com/river/msg/go/msg"
	fileCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_file"
	networkCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_network"
	queueCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_queue"
	syncCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_sync"
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

type contact struct {
	queueCtrl   *queueCtrl.Controller
	networkCtrl *networkCtrl.Controller
	fileCtrl    *fileCtrl.Controller
	syncCtrl    *syncCtrl.Controller
	sdk         module.SDK
}

func New() *contact {
	return &contact{}
}

func (r *contact) Init(sdk module.SDK) {
	r.sdk = sdk
	r.networkCtrl = sdk.NetCtrl()
	r.queueCtrl = sdk.QueueCtrl()
	r.syncCtrl = sdk.SyncCtrl()
	r.fileCtrl = sdk.FileCtrl()

}

func (r *contact) LocalHandlers() map[int64]domain.LocalMessageHandler {
	return map[int64]domain.LocalMessageHandler{
		msg.C_ContactsAdd:          r.contactsAdd,
		msg.C_ContactsDelete:       r.contactsDelete,
		msg.C_ContactsDeleteAll:    r.contactsDeleteAll,
		msg.C_ContactsGet:          r.contactsGet,
		msg.C_ContactsGetTopPeers:  r.contactsGetTopPeers,
		msg.C_ContactsImport:       r.contactsImport,
		msg.C_ContactsResetTopPeer: r.contactsResetTopPeer,
	}
}
