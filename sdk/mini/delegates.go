package mini

import (
	"git.ronaksoft.com/river/sdk/internal/request"
)

// MainDelegate external (UI) handler will listen to this function to receive data from SDK
type MainDelegate interface {
	OnNetworkStatusChanged(status int)
	OnSyncStatusChanged(status int)
	OnUpdates(constructor int64, b []byte)
	OnGeneralError(b []byte)
	OnSessionClosed(res int)
	ShowLoggerAlert()
	AddLog(text string)
	AppUpdate(version string, updateAvailable, force bool)
	DataSynced(dialogs, contacts, gifs bool)
}

type (
	RequestDelegateFlag = request.DelegateFlag
	RequestDelegate     = request.Delegate
)

// Request Flags
const (
	RequestServerForced RequestDelegateFlag = 1 << iota
	RequestBlocking
	RequestSkipWaitForNetwork
	RequestSkipFlusher
	RequestRealtime
	RequestBatch
)
