package domain

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/msg"
)

const (
	FilePayloadSize    = 1024 * 256        // 256KB
	FileMaxAllowedSize = 750 * 1024 * 1024 // 750MB
	FileMaxPhotoSize   = 1 * 1024 * 1024   // 1MB
	FileRetryThreshold = 10
	FilePipelineCount  = 8
)

// Global Parameters
const (
	PRODUCTION_SERVER_WEBSOCKET_ENDPOINT = "ws://river.im"
	DEFAULT_WS_PING_TIME                 = 10 * time.Second
	DEFAULT_WS_PONG_TIMEOUT              = 3 * time.Second
	DEFAULT_WS_WRITE_TIMEOUT             = 3 * time.Second
	DEFAULT_WS_REALTIME_TIMEOUT          = 3 * time.Second
	DEFAULT_REQUEST_TIMEOUT              = 30 * time.Second
	SnapshotSync_Threshold               = 999
)

// LOG KEYS
const (
	LK_FUNC_NAME        = "FUNC"
	LK_CLIENT_AUTH_ID   = "C_AUTHID"
	LK_SERVER_AUTH_ID   = "S_AUTHID"
	LK_MSG_KEY          = "MSG_KEY"
	LK_MSG_SIZE         = "MSG_SIZE"
	LK_DESC             = "DESC"
	LK_REQUEST_ID       = "REQUEST_ID"
	LK_CONSTRUCTOR_NAME = "CONSTRUCTOR"
)

// Table Column Names
const (
	CN_CONN_INFO    = "CONN_INFO"
	CN_UPDATE_ID    = "UPDATE_ID"
	CN_DEVICE_TOKEN = "DEVICE_TOKEN_INFO"
)

type NetworkStatus int

const (
	DISCONNECTED NetworkStatus = iota
	CONNECTING
	WEAK
	SLOW
	FAST
)

type SyncStatus int

const (
	OutOfSync SyncStatus = iota
	Syncing
	Synced
)

var SyncStatusName = map[SyncStatus]string{
	OutOfSync: "OutOfSync",
	Syncing:   "Syncing",
	Synced:    "Synced",
}
var NetworkStatusName = map[NetworkStatus]string{
	DISCONNECTED: "DISCONNECTED",
	CONNECTING:   "CONNECTING",
	WEAK:         "WEAK",
	SLOW:         "SLOW",
	FAST:         "FAST",
}

type RequestStatus int32

const (
	RequestStateNone       RequestStatus = 0
	RequestStateInProgress RequestStatus = 1
	RequestStateCompleted  RequestStatus = 2
	RequestStatePused      RequestStatus = 3
	RequestStateCanceled   RequestStatus = 4
	RequestStateError      RequestStatus = 5
)

var RequestStatusNames = map[RequestStatus]string{
	RequestStateNone:       "None",
	RequestStateInProgress: "InProgress",
	RequestStateCompleted:  "Completed",
	RequestStatePused:      "Pused",
	RequestStateCanceled:   "Canceled",
	RequestStateError:      "Error",
}

var DocumentAttributeTypeNames = map[msg.DocumentAttributeType]string{
	msg.AttributeTypeNone:  "AttributeTypeNone",
	msg.AttributeTypeAudio: "AttributeTypeAudio",
	msg.AttributeTypeVideo: "AttributeTypeVideo",
	msg.AttributeTypePhoto: "AttributeTypePhoto",
	msg.AttributeTypeFile:  "AttributeTypeFile",
	msg.AttributeAnimated:  "AttributeAnimated",
}
var MediaTypeNames = map[msg.MediaType]string{
	msg.MediaTypeEmpty:    "Empty",
	msg.MediaTypePhoto:    "Photo",
	msg.MediaTypeDocument: "Document",
	msg.MediaTypeContact:  "Contact",
}

type FileStateType int32

const (
	FileStateDownload             FileStateType = 1
	FileStateUpload               FileStateType = 2
	FileStateExistedDownload      FileStateType = 3
	FileStateExistedUpload        FileStateType = 4
	FileStateUploadAccountPhoto   FileStateType = 5
	FileStateDownloadAccountPhoto FileStateType = 6
	FileStateUploadGroupPhoto     FileStateType = 7
	FileStateDownloadGroupPhoto   FileStateType = 8
)

type MediaType int

const (
	MediaTypeAll   MediaType = 0
	MediaTypeFile  MediaType = 1
	MediaTypeMedia MediaType = 2
	MediaTypeVoice MediaType = 3
	MediaTypeAudio MediaType = 4
	MediaTypeLink  MediaType = 5
)
