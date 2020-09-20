package domain

import (
	"fmt"
	"git.ronaksoft.com/river/msg/msg"
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
	SysConfig      *msg.SystemConfig
)

// Global Parameters
const (
	DefaultWebsocketEndpoint = "ws://cyrus.river.im"
	WebsocketIdleTimeout     = 5 * time.Minute
	WebsocketPingTimeout     = 2 * time.Second
	WebsocketWriteTime       = 3 * time.Second
	WebsocketRequestTime     = 3 * time.Second
	WebsocketDialTimeout     = 3 * time.Second
	WebsocketDialTimeoutLong = 10 * time.Second
	HttpRequestTimeout       = 1 * time.Minute
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
	SkGifHash            = "GIF_HASH"
	SkTeam               = "TEAM"
)

func GetContactsGetHashKey(teamID int64) string {
	return fmt.Sprintf("%s.%021d", SkContactsGetHash, teamID)
}

// NetworkStatus network controller status
type NetworkStatus int

const (
	NetworkDisconnected NetworkStatus = 0
	NetworkConnecting   NetworkStatus = 1
	NetworkConnected    NetworkStatus = 4
)

func (ns NetworkStatus) ToString() string {
	switch ns {
	case NetworkDisconnected:
		return "Disconnected"
	case NetworkConnecting:
		return "Connecting"
	case NetworkConnected:
		return "Connected"
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

// UserStatus
const (
	LastSeenUnknown = iota
	LastSeenFewSeconds
	LastSeenFewMinutes
	LastSeenToday
	LastSeenYesterday
	LastSeenThisWeek
	LastSeenLastWeek
	LastSeenThisMonth
	LastSeenLongTimeAgo
)

// Network Connection
const (
	ConnectionNone = iota
	ConnectionWifi
	ConnectionCellular
)
