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

// OnFileStatusChanged delegate to rise event
type OnFileStatusChanged func(messageID, position, totalSize int64, stateType FileStateType)

// OnFileUploadCompleted delegate to rise event
type OnFileUploadCompleted func(messageID, fileID int64, clusterID, totalParts int32, stateType FileStateType, req *msg.ClientSendMessageMedia)

// OnFileDownloadCompleted delegate to rise event
type OnFileDownloadCompleted func(messageID int64, filePath string, stateType FileStateType)

// OnFileUploadError on receive error from server
type OnFileUploadError func(messageID, requestID int64, filePath string, err []byte)

// OnFileDownloadError on receive error from server
type OnFileDownloadError func(messageID, requestID int64, filePath string, err []byte)
