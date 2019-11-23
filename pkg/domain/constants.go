package domain

import (
	"time"
)

const (
	FilePayloadSize    = 1024 * 256        // 256KB
	FileMaxAllowedSize = 750 * 1024 * 1024 // 750MB
	FileMaxPhotoSize   = 1 * 1024 * 1024   // 1MB
	SDKVersion         = "v0.9.3"
)

var (
	ClientPlatform string
	ClientVersion  string
	ClientOS       string
	ClientVendor   string
)

// Global Parameters
const (
	WebsocketEndpoint    = "ws://cyrus.river.im"
	WebsocketWriteTime   = 3 * time.Second
	WebsocketRequestTime = 8 * time.Second
	WebsocketDialTimeout = 3 * time.Second
	HttpRequestTime      = 30 * time.Second

	SnapshotSyncThreshold = 10000
)

// System Keys
const (
	SkUpdateID           = "UPDATE_ID"
	SkDeviceToken        = "DEVICE_TOKEN_INFO"
	SkContactsImportHash = "CONTACTS_IMPORT_HASH"
	SkContactsGetHash    = "CONTACTS_GET_HASH"
	SkSystemSalts        = "SERVER_SALTS"
	SkReIndexTime        = "RE_INDEX_TS"
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

// UserStatus Times
const (
	Minute   = 60
	Hour     = Minute * 60
	Day      = Hour * 24
	Week     = Day * 7
	Month    = Week * 4
	TwoMonth = Month * 2
)
