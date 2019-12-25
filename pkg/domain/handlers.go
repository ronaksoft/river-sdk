package domain

import (
	msg "git.ronaksoftware.com/ronak/riversdk/msg/chat"
)

// ErrorHandler error callback/delegate
type ErrorHandler func(requestID uint64, u *msg.Error)

// MessageHandler success callback/delegate
type MessageHandler func(m *msg.MessageEnvelope)

// DeferredRequestHandler late responses that they requestID has been terminated callback/delegate
type DeferredRequestHandler func(constructor int64, msg []byte)

// OnUpdateMainDelegateHandler used as relay to pass getDifference updates to UI
type OnUpdateMainDelegateHandler func(constructor int64, msg []byte)

// OnConnectCallback networkController callback/delegate on websocket dial success
type OnConnectCallback func() error

// NetworkStatusUpdateCallback NetworkController status change callback/delegate
type NetworkStatusUpdateCallback func(newStatus NetworkStatus)

// SyncStatusUpdateCallback SyncController status change callback/delegate
type SyncStatusUpdateCallback func(newStatus SyncStatus)

// TimeoutCallback timeout callback/delegate
type TimeoutCallback func()

// UpdateApplier on receive update in SyncController, cache client data, there are some applier function for each proto message
type UpdateApplier func(envelope *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error)

// MessageApplier on receive response in SyncController, cache client data, there are some applier function for each proto message
type MessageApplier func(envelope *msg.MessageEnvelope)

// LocalMessageHandler SDK commands that handle user request from client cache
type LocalMessageHandler func(in, out *msg.MessageEnvelope, timeoutCB TimeoutCallback, successCB MessageHandler)

// ReceivedMessageHandler NetworkController pass all received response messages to this callback/delegate
type ReceivedMessageHandler func(messages []*msg.MessageEnvelope)

// ReceivedUpdateHandler NetworkController pass all received update messages to this callback/delegate
type ReceivedUpdateHandler func(container *msg.UpdateContainer)
