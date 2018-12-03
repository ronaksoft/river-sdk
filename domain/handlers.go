package domain

import (
	"git.ronaksoftware.com/ronak/riversdk/msg"
)

// ErrorHandler
type ErrorHandler func(u *msg.Error)

// // UpdateHandler
// type UpdateHandler func(u *msg.UpdateContainer)

// MessageHandler
type MessageHandler func(m *msg.MessageEnvelope)

// DeferredRequestHandler
type DeferredRequestHandler func(constructor int64, msg []byte)

// OnUpdateMainDelegateHandler
type OnUpdateMainDelegateHandler func(constructor int64, msg []byte)

// OnConnectCallback
type OnConnectCallback func()

// NetworkStatusUpdateCallback
type NetworkStatusUpdateCallback func(newStatus NetworkStatus)

// SyncStatusUpdateCallback
type SyncStatusUpdateCallback func(newStatus SyncStatus)

// TimeoutCallback
type TimeoutCallback func()

// UpdateApplier
type UpdateApplier func(envelope *msg.UpdateEnvelope) []*msg.UpdateEnvelope

// MessageApplier
type MessageApplier func(envelope *msg.MessageEnvelope)

// LocalMessageHandler
type LocalMessageHandler func(in, out *msg.MessageEnvelope, timeoutCB TimeoutCallback, successCB MessageHandler)

// OnMessageHandler
type OnMessageHandler func(messages []*msg.MessageEnvelope)

// OnUpdateHandler
type OnUpdateHandler func(messages []*msg.UpdateContainer)
