package domain

import (
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
)

// ErrorHandler error callback/delegate
type ErrorHandler func(u *msg.Error)

// MessageHandler success callback/delegate
type MessageHandler func(m *msg.MessageEnvelope)

// DeferredRequestHandler late responses that they requestID has been terminated callback/delegate
type DeferredRequestHandler func(constructor int64, msg []byte)

// OnUpdateMainDelegateHandler used as relay to pass getDifference updates to UI
type OnUpdateMainDelegateHandler func(constructor int64, msg []byte)

// OnConnectCallback networkController callback/delegate on websocket dial success
type OnConnectCallback func()

// NetworkStatusUpdateCallback NetworkController status change callback/delegate
type NetworkStatusUpdateCallback func(newStatus NetworkStatus)

// SyncStatusUpdateCallback SyncController status change callback/delegate
type SyncStatusUpdateCallback func(newStatus SyncStatus)

// TimeoutCallback timeout callback/delegate
type TimeoutCallback func()

// UpdateApplier on receive update in SyncController, cache client data, there are some applier function for each proto message
type UpdateApplier func(envelope *msg.UpdateEnvelope) []*msg.UpdateEnvelope

// MessageApplier on receive response in SyncController, cache client data, there are some applier function for each proto message
type MessageApplier func(envelope *msg.MessageEnvelope)

// LocalMessageHandler SDK commands that handle user request from client cache
type LocalMessageHandler func(in, out *msg.MessageEnvelope, timeoutCB TimeoutCallback, successCB MessageHandler)

// ReceivedMessageHandler NetworkController pass all received response messages to this callback/delegate
type ReceivedMessageHandler func(messages []*msg.MessageEnvelope)

// ReceivedUpdateHandler NetworkController pass all received update messages to this callback/delegate
type ReceivedUpdateHandler func(messages []*msg.UpdateContainer)

// OnFileStatusChanged delegate to rise file progress event
type OnFileStatusChanged func(messageID, processedParts, totalParts int64, stateType FileStateType)

// OnFileUploadCompleted delegate to rise upload completed event
type OnFileUploadCompleted func(messageID, fileID, targetID int64,
	clusterID int32, totalParts int64,
	stateType FileStateType,
	filePath string,
	req *msg.ClientSendMessageMedia,
	thumbFileID int64,
	thumbTotalParts int32,
)

// OnFileDownloadCompleted delegate to rise download completed event
type OnFileDownloadCompleted func(messageID int64, filePath string, stateType FileStateType)

// OnFileUploadError on receive error from server
type OnFileUploadError func(messageID, requestID int64, filePath string, err []byte)

// OnFileDownloadError on receive error from server
type OnFileDownloadError func(messageID, requestID int64, filePath string, err []byte)
