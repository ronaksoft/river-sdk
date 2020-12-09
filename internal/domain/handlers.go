package domain

import (
	"git.ronaksoft.com/river/msg/go/msg"
)

// DeferredRequestHandler late responses that they requestID has been terminated callback/delegate
type DeferredRequestHandler func(constructor int64, msg []byte)

// UpdateReceivedCallback used as relay to pass getDifference updates to UI
type UpdateReceivedCallback func(constructor int64, msg []byte)

// AppUpdateCallback will be called to inform client of any update available
type AppUpdateCallback func(version string, updateAvailable bool, force bool)

// OnConnectCallback networkController callback/delegate on websocket dial success
type OnConnectCallback func() error

// NetworkStatusChangeCallback NetworkController status change callback/delegate
type NetworkStatusChangeCallback func(newStatus NetworkStatus)

// SyncStatusChangeCallback SyncController status change callback/delegate
type SyncStatusChangeCallback func(newStatus SyncStatus)

// TimeoutCallback timeout callback/delegate
type TimeoutCallback func()

// ErrorHandler error callback/delegate
type ErrorHandler func(requestID uint64, u *msg.Error)

// MessageHandler success callback/delegate
type MessageHandler func(m *msg.MessageEnvelope)

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
