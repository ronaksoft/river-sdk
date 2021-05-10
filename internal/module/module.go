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

type SDK interface {
	Version() string
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

type Base struct {
	queueCtrl   *queueCtrl.Controller
	networkCtrl *networkCtrl.Controller
	fileCtrl    *fileCtrl.Controller
	syncCtrl    *syncCtrl.Controller
	sdk         SDK
	handlers    map[int64]domain.LocalMessageHandler
}

func (b Base) Init(sdk SDK) {
	b.sdk = sdk
}

func (b Base) Execute(in *rony.MessageEnvelope, onTimeout domain.TimeoutCallback, onComplete domain.MessageHandler) {
	out := &rony.MessageEnvelope{}
	h := b.handlers[in.Constructor]
	if h == nil {
		out.Fill(in.RequestID, rony.C_Error, &rony.Error{Code: "E100", Items: "MODULE_HANDLER_NOT_FOUND"})
		onComplete(out)
		return
	}
	h(in, out, onTimeout, onComplete)
}

func (b Base) RegisterHandlers(handlers map[int64]domain.LocalMessageHandler) {
	b.handlers = handlers
}

func (b Base) LocalHandlers() map[int64]domain.LocalMessageHandler {
	return b.handlers
}

func (b Base) SDK() SDK {
	return b.sdk
}
