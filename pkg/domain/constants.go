package domain

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/msg"
)

const (
	FilePayloadSize    = 1024 * 256        // 256KB
	FileMaxAllowedSize = 750 * 1024 * 1024 // 750MB
	FileMaxPhotoSize   = 1 * 1024 * 1024   // 1MB
	FileMaxRetry       = 10
	FilePipelineCount  = 8
)

// Global Parameters
const (
	// WebsocketEndpoint production server address
	WebsocketEndpoint     = "ws://river.im"
	WebsocketPingTime     = 10 * time.Second
	WebsocketPongTime     = 3 * time.Second
	WebsocketWriteTime    = 3 * time.Second
	WebsocketDirectTime   = 3 * time.Second
	WebsocketRequestTime  = 30 * time.Second
	SnapshotSyncThreshold = 10000
)

// Table Column Names
const (
	ColumnConnectionInfo     = "CONN_INFO"
	ColumnUpdateID           = "UPDATE_ID"
	ColumnDeviceToken        = "DEVICE_TOKEN_INFO"
	ColumnContactsImportHash = "CONTACTS_IMPORT_HASH"
	ColumnContactsGetHash    = "CONTACTS_GET_HASH"
)

// NetworkStatus network controller status
type NetworkStatus int

const (
	// NetworkDisconnected no internet
	NetworkDisconnected NetworkStatus = iota
	// NetworkConnecting websocket dialing
	NetworkConnecting
	// NetworkWeak weak
	NetworkWeak
	// NetworkSlow slow
	NetworkSlow
	// NetworkFast fast
	NetworkFast
)

func (ns NetworkStatus) ToString() string {
	switch ns {
	case NetworkDisconnected:
		return "Disconnected"
	case NetworkConnecting:
		return "Connecting"
	case NetworkWeak:
		return "Weak"
	case NetworkSlow:
		return "Slow"
	case NetworkFast:
		return "Fast"
	}
	return "Unknown"
}

// SyncStatus status of synchronizer
type SyncStatus int

const (
	// OutOfSync synchronizer is fall behind
	OutOfSync SyncStatus = iota
	// Syncing synchronizer is running sync request
	Syncing
	// Synced synchronizer finished sync request
	Synced
)

func (ss SyncStatus) ToString() string {
	switch ss {
	case OutOfSync:
		return "OutOfSync"
	case Syncing:
		return "Syncing"
	case Synced:
		return "Synced"
	}
	return ""
}

// RequestStatus state of file download/upload request
type RequestStatus int32

const (
	// RequestStateNone no request invoked
	RequestStateNone RequestStatus = 0
	// RequestStateInProgress downloading/uploading
	RequestStateInProgress RequestStatus = 1
	// RequestStateCompleted already file is downloaded/uploaded
	RequestStateCompleted RequestStatus = 2
	// RequestStatePaused paused
	RequestStatePaused RequestStatus = 3
	// RequestStateCanceled canceled by user
	RequestStateCanceled RequestStatus = 4
	// RequestStateError encountered error
	RequestStateError RequestStatus = 5
)

func (rs RequestStatus) ToString() string {
	switch rs {
	case RequestStateNone:
		return "None"
	case RequestStateInProgress:
		return "InProgress"
	case RequestStateCompleted:
		return "Completed"
	case RequestStatePaused:
		return "Paused"
	case RequestStateCanceled:
		return "Canceled"
	case RequestStateError:
		return "Error"
	}
	return ""
}

// DocumentAttributeTypeNames log helper to retrive DocumentAttributeType name
var DocumentAttributeTypeNames = map[msg.DocumentAttributeType]string{
	msg.AttributeTypeNone:  "AttributeTypeNone",
	msg.AttributeTypeAudio: "AttributeTypeAudio",
	msg.AttributeTypeVideo: "AttributeTypeVideo",
	msg.AttributeTypePhoto: "AttributeTypePhoto",
	msg.AttributeTypeFile:  "AttributeTypeFile",
	msg.AttributeAnimated:  "AttributeAnimated",
}

// MediaTypeNames log helper to retrive SharedMediaType name
var MediaTypeNames = map[msg.MediaType]string{
	msg.MediaTypeEmpty:    "Empty",
	msg.MediaTypePhoto:    "Photo",
	msg.MediaTypeDocument: "Document",
	msg.MediaTypeContact:  "Contact",
}

// FileStateType provide some status to distinguish file upload/download and its progress types
type FileStateType int32

const (
	// FileStateDownload download
	FileStateDownload FileStateType = 1
	// FileStateUpload upload
	FileStateUpload FileStateType = 2
	// FileStateExistedDownload file already exist
	FileStateExistedDownload FileStateType = 3
	// FileStateExistedUpload file uploaded document already exist
	FileStateExistedUpload FileStateType = 4
	// FileStateUploadAccountPhoto the request AccountUploadPhoto
	FileStateUploadAccountPhoto FileStateType = 5
	// FileStateDownloadAccountPhoto the request AccountDownloadPhoto
	FileStateDownloadAccountPhoto FileStateType = 6
	// FileStateUploadGroupPhoto the request GroupUploadPhoto
	FileStateUploadGroupPhoto FileStateType = 7
	// FileStateDownloadGroupPhoto the request GroupDownloadPhoto
	FileStateDownloadGroupPhoto FileStateType = 8
)

// SharedMediaType filter for displaying shared medias
type SharedMediaType int

const (
	// SharedMediaTypeAll all documents
	SharedMediaTypeAll SharedMediaType = 0
	// SharedMediaTypeFile files
	SharedMediaTypeFile SharedMediaType = 1
	// SharedMediaTypeMedia photo/video/animated
	SharedMediaTypeMedia SharedMediaType = 2
	// SharedMediaTypeVoice audio document that have IsVoice flag
	SharedMediaTypeVoice SharedMediaType = 3
	// SharedMediaTypeAudio audio document that its IsVoice flag is flase
	SharedMediaTypeAudio SharedMediaType = 4
	// SharedMediaTypeLink displays all messages that have link entity
	SharedMediaTypeLink SharedMediaType = 5
)
