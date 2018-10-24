package domain

import "time"

// Global Parameters
const (
	PRODUCTION_SERVER_WEBSOCKET_ENDPOINT = "ws://river.im"
	DEFAULT_WS_PING_TIME                 = 10 * time.Second
	DEFAULT_WS_PONG_TIMEOUT              = 3 * time.Second
	DEFAULT_WS_WRITE_TIMEOUT             = 3 * time.Second
	DEFAULT_WS_REALTIME_TIMEOUT          = 3 * time.Second
	DEFAULT_REQUEST_TIMEOUT              = 30 * time.Second
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
	CN_CONN_INFO = "CONN_INFO"
	CN_UPDATE_ID = "UPDATE_ID"
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
