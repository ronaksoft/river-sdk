package domain

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/msg"
)

const (
	FilePayloadSize    = 1024 * 256        // 256KB
	FileMaxAllowedSize = 750 * 1024 * 1024 // 750MB
)

// Global Parameters
const (
	PRODUCTION_SERVER_WEBSOCKET_ENDPOINT = "ws://river.im"
	DEFAULT_WS_PING_TIME                 = 10 * time.Second
	DEFAULT_WS_PONG_TIMEOUT              = 3 * time.Second
	DEFAULT_WS_WRITE_TIMEOUT             = 3 * time.Second
	DEFAULT_WS_REALTIME_TIMEOUT          = 3 * time.Second
	DEFAULT_REQUEST_TIMEOUT              = 30 * time.Second
	SnapshotSync_Threshold               = 200
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
	RequestStateDefault   RequestStatus = 0
	RequestStateCompleted RequestStatus = 1
	RequestStatePending   RequestStatus = 2
	RequestStatePused     RequestStatus = 3
	RequestStateFailed    RequestStatus = 4
	RequestStateCanceled  RequestStatus = 5
)

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
