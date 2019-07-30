package domain

import (
	"time"
)

const (
	FilePayloadSize    = 1024 * 256        // 256KB
	FileMaxAllowedSize = 750 * 1024 * 1024 // 750MB
	FileMaxPhotoSize   = 1 * 1024 * 1024   // 1MB
	FileMaxRetry       = 10
	FilePipelineCount  = 8
	SDKVersion = "v0.8.1"
)

// Global Parameters
const (
	WebsocketEndpoint     = "ws://river.im"
	WebsocketPingTime     = 30 * time.Second
	WebsocketPongTime     = 10 * time.Second
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
	NetworkDisconnected NetworkStatus = iota
	NetworkConnecting
	NetworkWeak
	NetworkSlow
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
	RequestStatusNone       RequestStatus = 0 // RequestStatusNone no request invoked
	RequestStatusInProgress RequestStatus = 1 // RequestStatusInProgress downloading/uploading
	RequestStatusCompleted  RequestStatus = 2 // RequestStatusCompleted already file is downloaded/uploaded
	RequestStatusPaused     RequestStatus = 3 // RequestStatusPaused paused
	RequestStatusCanceled   RequestStatus = 4 // RequestStatusCanceled canceled by user
	RequestStatusError      RequestStatus = 5 // RequestStatusError encountered error
)

func (rs RequestStatus) ToString() string {
	switch rs {
	case RequestStatusNone:
		return "None"
	case RequestStatusInProgress:
		return "InProgress"
	case RequestStatusCompleted:
		return "Completed"
	case RequestStatusPaused:
		return "Paused"
	case RequestStatusCanceled:
		return "Canceled"
	case RequestStatusError:
		return "Error"
	}
	return ""
}

// FileStateType provide some status to distinguish file upload/download and its progress types
type FileStateType int32

const (
	FileStateDownload           FileStateType = 1 // FileStateDownload download
	FileStateUpload             FileStateType = 2 // FileStateUpload upload
	FileStateExistedDownload    FileStateType = 3 // FileStateExistedDownload file already exist
	FileStateExistedUpload      FileStateType = 4 // FileStateExistedUpload file uploaded document already exist
	FileStateUploadAccountPhoto FileStateType = 5 // FileStateUploadAccountPhoto the request AccountUploadPhoto
	FileStateUploadGroupPhoto   FileStateType = 7 // FileStateUploadGroupPhoto the request GroupUploadPhoto
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
