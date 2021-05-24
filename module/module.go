package module

import (
	fileCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_file"
	networkCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_network"
	queueCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_queue"
	syncCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_sync"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"github.com/ronaksoft/rony"
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

// SDK represents the Prime or Mini River.
type SDK interface {
	Version() string
	SyncCtrl() *syncCtrl.Controller
	NetCtrl() *networkCtrl.Controller
	QueueCtrl() *queueCtrl.Controller
	FileCtrl() *fileCtrl.Controller
	GetConnInfo() domain.RiverConfigurator
	Module(name string) Module
	Execute(constructor int64, commandBytes []byte, cb domain.Callback, flags domain.RequestDelegateFlag) (requestID int64, err error)
	ExecuteWithTeam(teamID, accessHash, constructor int64, commandBytes []byte, cb domain.Callback, flags domain.RequestDelegateFlag, msTimeout int64) (requestID int64, err error)
}

type Module interface {
	Name() string
	Init(sdk SDK)
	LocalHandlers() map[int64]domain.LocalHandler
	UpdateAppliers() map[int64]domain.UpdateApplier
	MessageAppliers() map[int64]domain.MessageApplier
	Execute(in *rony.MessageEnvelope, da domain.Callback)
}

// Base provides the boilerplate code for every module. Hence developer only needs to write the module specific
// LocalHandlers, UpdateAppliers and MessageAppliers.
type Base struct {
	queueCtrl       *queueCtrl.Controller
	networkCtrl     *networkCtrl.Controller
	fileCtrl        *fileCtrl.Controller
	syncCtrl        *syncCtrl.Controller
	sdk             SDK
	handlers        map[int64]domain.LocalHandler
	updateAppliers  map[int64]domain.UpdateApplier
	messageAppliers map[int64]domain.MessageApplier
}

func (b *Base) Init(sdk SDK) {
	b.sdk = sdk
}

func (b *Base) Execute(in *rony.MessageEnvelope, da domain.Callback) {
	out := &rony.MessageEnvelope{}
	h := b.handlers[in.Constructor]
	if h == nil {
		out.Fill(in.RequestID, rony.C_Error, &rony.Error{Code: "E100", Items: "MODULE_HANDLER_NOT_FOUND"})
		da.OnComplete(out)
		return
	}

	h(in, out, da)
}

func (b *Base) RegisterUpdateAppliers(appliers map[int64]domain.UpdateApplier) {
	b.updateAppliers = appliers
}

func (b *Base) UpdateAppliers() map[int64]domain.UpdateApplier {
	return b.updateAppliers
}

func (b *Base) RegisterMessageAppliers(appliers map[int64]domain.MessageApplier) {
	b.messageAppliers = appliers
}

func (b *Base) MessageAppliers() map[int64]domain.MessageApplier {
	return b.messageAppliers
}

func (b *Base) RegisterHandlers(handlers map[int64]domain.LocalHandler) {
	b.handlers = handlers
}

func (b *Base) LocalHandlers() map[int64]domain.LocalHandler {
	return b.handlers
}

func (b *Base) SDK() SDK {
	return b.sdk
}
