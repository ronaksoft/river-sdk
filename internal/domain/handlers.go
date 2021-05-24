package domain

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/pools"
)

type RequestDelegateFlag = int32

// Request Flags
const (
	RequestServerForced RequestDelegateFlag = 1 << iota
	RequestBlocking
	RequestSkipWaitForNetwork
	RequestSkipFlusher
	RequestRealtime
	RequestBatch
)

type RequestDelegate interface {
	OnComplete(b []byte)
	OnTimeout(err error)
	OnProgress(percent int64)
	Flags() RequestDelegateFlag
}

type delegateAdapter struct {
	d  RequestDelegate
	ui bool
}

func DelegateAdapterFromRequest(d RequestDelegate, ui bool) *delegateAdapter {
	return &delegateAdapter{
		d:  d,
		ui: ui,
	}
}

func (rda *delegateAdapter) OnComplete(m *rony.MessageEnvelope) {
	if rda.d == nil {
		return
	}
	buf := pools.Buffer.FromProto(m)
	rda.d.OnComplete(*buf.Bytes())
	pools.Buffer.Put(buf)
}

func (rda *delegateAdapter) OnTimeout() {
	if rda.d == nil {
		return
	}
	rda.d.OnTimeout(ErrRequestTimeout)
}

func (rda *delegateAdapter) OnProgress(percent int64) {
	if rda.d == nil {
		return
	}
	rda.d.OnProgress(percent)
}

func (rda *delegateAdapter) UI() bool {
	return rda.ui
}

type Callback interface {
	OnComplete(m *rony.MessageEnvelope)
	OnTimeout()
	OnProgress(percent int64)
	UI() bool
}

type callback struct {
	onComplete func(m *rony.MessageEnvelope)
	onTimeout  func()
	onProgress func(percent int64)
	ui         bool
}

func (c *callback) OnComplete(m *rony.MessageEnvelope) {
	if c.onComplete == nil {
		return
	}
	c.onComplete(m)
}

func (c *callback) OnTimeout() {
	if c.onTimeout == nil {
		return
	}
	c.onTimeout()
}

func (c *callback) OnProgress(percent int64) {
	if c.onProgress == nil {
		return
	}
	c.onProgress(percent)
}

func (c *callback) UI() bool {
	return c.ui
}

func NewCallback(onTimeout func(), onComplete func(envelope *rony.MessageEnvelope), onProgress func(int64), ui bool) *callback {
	return &callback{
		onComplete: onComplete,
		onTimeout:  onTimeout,
		onProgress: onProgress,
		ui:         ui,
	}
}

func EmptyCallback() *callback {
	return &callback{}
}

type requestDelegate struct {
	onComplete func(b []byte)
	onTimeout  func(error)
	onProgress func(int64)
	flags      RequestDelegateFlag
}

func (r *requestDelegate) OnComplete(b []byte) {
	if r.onComplete != nil {
		r.onComplete(b)
	}
}

func (r *requestDelegate) OnTimeout(err error) {
	if r.onTimeout != nil {
		r.onTimeout(err)
	}
}

func (r *requestDelegate) Flags() RequestDelegateFlag {
	return r.flags
}

func (r *requestDelegate) OnProgress(percent int64) {
	if r.onProgress != nil {
		r.onProgress(percent)
	}
}

func NewRequestDelegate(onComplete func(b []byte), onTimeout func(err error), flags RequestDelegateFlag) *requestDelegate {
	return &requestDelegate{
		onComplete: onComplete,
		onTimeout:  onTimeout,
		onProgress: nil,
		flags:      flags,
	}
}

// UpdateReceivedCallback used as relay to pass getDifference updates to UI
type UpdateReceivedCallback func(constructor int64, msg []byte)

// AppUpdateCallback will be called to inform client of any update available
type AppUpdateCallback func(version string, updateAvailable bool, force bool)

// DataSyncedCallback will be called to inform client of update of synced data
type DataSyncedCallback func(dialogs, contacts, gifs bool)

// OnConnectCallback networkController callback/delegate on websocket dial success
type OnConnectCallback func() error

// NetworkStatusChangeCallback NetworkController status change callback/delegate
type NetworkStatusChangeCallback func(newStatus NetworkStatus)

// SyncStatusChangeCallback SyncController status change callback/delegate
type SyncStatusChangeCallback func(newStatus SyncStatus)

// TimeoutCallback timeout callback/delegate
type TimeoutCallback func()

// ErrorHandler error callback/delegate
type ErrorHandler func(requestID uint64, u *rony.Error)

// MessageHandler success callback/delegate
type MessageHandler func(m *rony.MessageEnvelope)

// UpdateApplier on receive update in SyncController, cache client data, there are some applier function for each proto message
type UpdateApplier func(envelope *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error)

// MessageApplier on receive response in SyncController, cache client data, there are some applier function for each proto message
type MessageApplier func(envelope *rony.MessageEnvelope)

// LocalHandler SDK commands that handle user request from client cache
type LocalHandler func(in, out *rony.MessageEnvelope, da Callback)

// ReceivedMessageHandler NetworkController pass all received response messages to this callback/delegate
type ReceivedMessageHandler func(messages []*rony.MessageEnvelope)

// ReceivedUpdateHandler NetworkController pass all received update messages to this callback/delegate
type ReceivedUpdateHandler func(container *msg.UpdateContainer)
