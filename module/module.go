package module

import (
	fileCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_file"
	networkCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_network"
	queueCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_queue"
	syncCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_sync"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/request"
	"github.com/ronaksoft/rony/errors"
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
	Execute(cb request.Callback) (err error)
}

type Module interface {
	Name() string
	Init(sdk SDK, logger *logs.Logger)
	LocalHandlers() map[int64]request.LocalHandler
	UpdateAppliers() map[int64]domain.UpdateApplier
	MessageAppliers() map[int64]domain.MessageApplier
	Execute(da request.Callback)
}

// Base provides the boilerplate code for every module. Hence developer only needs to write the module specific
// LocalHandlers, UpdateAppliers and MessageAppliers.
type Base struct {
	sdk             SDK
	handlers        map[int64]request.LocalHandler
	updateAppliers  map[int64]domain.UpdateApplier
	messageAppliers map[int64]domain.MessageApplier
	logger          *logs.Logger
}

func (b *Base) Init(sdk SDK, logger *logs.Logger) {
	b.sdk = sdk
	b.logger = logger
}

func (b *Base) Execute(da request.Callback) {
	h := b.handlers[da.Constructor()]
	if h == nil {
		da.OnComplete(errors.Message(da.RequestID(), "E100", "MODULE_HANDLER_NOT_FOUND"))
		return
	}

	h(da)
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

func (b *Base) RegisterHandlers(handlers map[int64]request.LocalHandler) {
	b.handlers = handlers
}

func (b *Base) LocalHandlers() map[int64]request.LocalHandler {
	return b.handlers
}

func (b *Base) SDK() SDK {
	return b.sdk
}

func (b *Base) Log() *logs.Logger {
	return b.logger
}
