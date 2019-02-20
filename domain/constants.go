package domain

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/msg"
)

const (
	// FilePayloadSize file upload chunk size
	FilePayloadSize = 1024 * 256 // 256KB
	// FileMaxAllowedSize max file size limit
	FileMaxAllowedSize = 750 * 1024 * 1024 // 750MB
	// FileMaxPhotoSize max allowed photo size
	FileMaxPhotoSize = 1 * 1024 * 1024 // 1MB
	// FileRetryThreshold downloader/uploader stops retry after this threshold
	FileRetryThreshold = 10
	// FilePipelineCount multiconnection download/upload
	FilePipelineCount = 8
)

// Global Parameters
const (
	// WebsocketEndpoint production server address
	WebsocketEndpoint = "ws://river.im"
	// WebsocketPingTime ping timeout
	WebsocketPingTime = 10 * time.Second
	// WebsocketPongTime pong timeout
	WebsocketPongTime = 3 * time.Second
	// WebsocketWriteTime send request timeout
	WebsocketWriteTime = 3 * time.Second
	// WebsocketDirectTime send realtime request timeout
	WebsocketDirectTime = 3 * time.Second
	// WebsocketRequestTime wait for response timeout
	WebsocketRequestTime = 30 * time.Second
	// SnapshotSyncThreshold snapshot threshold
	SnapshotSyncThreshold = 999
)

// Table Column Names
const (
	ColumnConnectionInfo = "CONN_INFO"
	ColumnUpdateID       = "UPDATE_ID"
	ColumnDeviceToken    = "DEVICE_TOKEN_INFO"
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

// SyncStatusName log helper to retrive SyncStatus name
var SyncStatusName = map[SyncStatus]string{
	OutOfSync: "OutOfSync",
	Syncing:   "Syncing",
	Synced:    "Synced",
}

// NetworkStatusName log helper to retrive NetworkStatus name
var NetworkStatusName = map[NetworkStatus]string{
	NetworkDisconnected: "NetworkDisconnected",
	NetworkConnecting:   "NetworkConnecting",
	NetworkWeak:         "NetworkWeak",
	NetworkSlow:         "NetworkSlow",
	NetworkFast:         "NetworkFast",
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

// RequestStatusNames log helper to retrive RequestStatus name
var RequestStatusNames = map[RequestStatus]string{
	RequestStateNone:       "None",
	RequestStateInProgress: "InProgress",
	RequestStateCompleted:  "Completed",
	RequestStatePaused:     "Paused",
	RequestStateCanceled:   "Canceled",
	RequestStateError:      "Error",
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
