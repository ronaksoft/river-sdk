package module

import (
	fileCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_file"
	networkCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_network"
	queueCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_queue"
	syncCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_sync"
	"git.ronaksoft.com/river/sdk/internal/domain"
)

/*
   Creation Time: 2021 - May - 10
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

// RequestDelegate each request should have this callbacks
type RequestDelegate interface {
	OnComplete(b []byte)
	OnTimeout(err error)
	Flags() int32
}

type SDK interface {
	Version() string
	// ExecuteCommand(constructor int64, commandBytes []byte, delegate RequestDelegate) (requestID int64, err error)
	// ExecuteCommandWithTeam(teamID, accessHash, constructor int64, commandBytes []byte, delegate RequestDelegate) (requestID int64, err error)
	SyncCtrl() *syncCtrl.Controller
	NetCtrl() *networkCtrl.Controller
	QueueCtrl() *queueCtrl.Controller
	FileCtrl() *fileCtrl.Controller
	GetConnInfo() domain.RiverConfigurator
}

type Module interface {
	Init(sdk SDK)
	LocalHandlers() map[int64]domain.LocalMessageHandler
}
