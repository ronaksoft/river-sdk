package domain

import (
	"time"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
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
	ColumnSystemSalts        = "SERVER_SALTS"
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

// SyncStatus status of Sync Controller
type SyncStatus int

const (
	OutOfSync SyncStatus = iota
	Syncing
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
	RequestStateNone       RequestStatus = 0 // RequestStateNone no request invoked
	RequestStateInProgress RequestStatus = 1 // RequestStateInProgress downloading/uploading
	RequestStateCompleted  RequestStatus = 2 // RequestStateCompleted already file is downloaded/uploaded
	RequestStatePaused     RequestStatus = 3 // RequestStatePaused paused
	RequestStateCanceled   RequestStatus = 4 // RequestStateCanceled canceled by user
	RequestStateError      RequestStatus = 5 // RequestStateError encountered error
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

// MediaTypeNames log helper to retrieve SharedMediaType name
var MediaTypeNames = map[msg.MediaType]string{
	msg.MediaTypeEmpty:    "Empty",
	msg.MediaTypePhoto:    "Photo",
	msg.MediaTypeDocument: "Document",
	msg.MediaTypeContact:  "Contact",
}

// FileStateType provide some status to distinguish file upload/download and its progress types
type FileStateType int32

const (
	FileStateDownload             FileStateType = 1 // FileStateDownload download
	FileStateUpload               FileStateType = 2 // FileStateUpload upload
	FileStateExistedDownload      FileStateType = 3 // FileStateExistedDownload file already exist
	FileStateExistedUpload        FileStateType = 4 // FileStateExistedUpload file uploaded document already exist
	FileStateUploadAccountPhoto   FileStateType = 5 // FileStateUploadAccountPhoto the request AccountUploadPhoto
	FileStateDownloadAccountPhoto FileStateType = 6 // FileStateDownloadAccountPhoto the request AccountDownloadPhoto
	FileStateUploadGroupPhoto     FileStateType = 7 // FileStateUploadGroupPhoto the request GroupUploadPhoto
	FileStateDownloadGroupPhoto   FileStateType = 8 // FileStateDownloadGroupPhoto the request GroupDownloadPhoto
)

// SharedMediaType filter for displaying shared medias
type SharedMediaType int

const (
	SharedMediaTypeAll   SharedMediaType = 0 // SharedMediaTypeAll all documents
	SharedMediaTypeFile  SharedMediaType = 1 // SharedMediaTypeFile files
	SharedMediaTypeMedia SharedMediaType = 2 // SharedMediaTypeMedia photo/video/animated
	SharedMediaTypeVoice SharedMediaType = 3 // SharedMediaTypeVoice audio document that have IsVoice flag
	SharedMediaTypeAudio SharedMediaType = 4 // SharedMediaTypeAudio audio document that its IsVoice flag is false
	SharedMediaTypeLink  SharedMediaType = 5 // SharedMediaTypeLink displays all messages that have link entity
)
