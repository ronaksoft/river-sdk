package shared

import "time"

const (
	// DefaultServerURL websocket url
	DefaultServerURL = "ws://new.river.im"

	// DefaultTimeout request timeout
	DefaultTimeout = 10 * time.Second

	// DefaultSendTimeout write to ws timeout
	DefaultSendTimeout = 3 * time.Second
)
