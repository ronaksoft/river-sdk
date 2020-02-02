package domain

import (
	"time"
)

//go:generate go run update_version.go
const (
	FilePayloadSize = 1024 * 256 // 256KB
)

var (
	ClientPlatform string
	ClientVersion  string
	ClientOS       string
	ClientVendor   string
	ClientPhone    string
)

// Global Parameters
const (
	WebsocketEndpoint        = "ws://cyrus.river.im"
	WebsocketWriteTime       = 3 * time.Second
	WebsocketRequestTime     = 8 * time.Second
	WebsocketDialTimeout     = 3 * time.Second
	WebsocketDialTimeoutLong = 10 * time.Second
	HttpRequestTime          = 2 * time.Minute
	SnapshotSyncThreshold    = 10000
)

// System Keys
const (
	SkUpdateID           = "UPDATE_ID"
	SkDeviceToken        = "DEVICE_TOKEN_INFO"
	SkContactsImportHash = "CONTACTS_IMPORT_HASH"
	SkContactsGetHash    = "CONTACTS_GET_HASH"
	SkSystemSalts        = "SERVER_SALTS"
	SkReIndexTime        = "RE_INDEX_TS"
	SkLabelMinID         = "LABEL_MIN_ID"
	SkLabelMaxID         = "LABEL_MAX_ID"
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
	case RequestStatusCanceled:
		return "Canceled"
	case RequestStatusError:
		return "Error"
	}
	return ""
}

// UserStatus Times
const (
	Minute   = 60
	Hour     = Minute * 60
	Day      = Hour * 24
	Week     = Day * 7
	Month    = Week * 4
	TwoMonth = Month * 2
)
