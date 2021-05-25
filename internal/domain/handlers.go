package domain

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"github.com/ronaksoft/rony"
)

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
