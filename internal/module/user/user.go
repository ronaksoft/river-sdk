package user

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

type user struct {
	queueCtrl   *queueCtrl.Controller
	networkCtrl *networkCtrl.Controller
	fileCtrl    *fileCtrl.Controller
	syncCtrl    *syncCtrl.Controller
	sdk         module.SDK
}

func New() *user {
	return &user{}
}

func (r *user) Init(sdk module.SDK) {
	r.sdk = sdk
	r.networkCtrl = sdk.NetCtrl()
	r.queueCtrl = sdk.QueueCtrl()
	r.syncCtrl = sdk.SyncCtrl()
	r.fileCtrl = sdk.FileCtrl()

}

func (r *user) LocalHandlers() map[int64]domain.LocalMessageHandler {
	return map[int64]domain.LocalMessageHandler{
		msg.C_UsersGet:     r.usersGet,
		msg.C_UsersGetFull: r.usersGetFull,
	}
}
