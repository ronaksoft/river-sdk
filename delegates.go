package riversdk

// MainDelegate external (UI) handler will listen to this function to receive data from SDK
type MainDelegate interface {
	OnNetworkStatusChanged(status int)
	OnSyncStatusChanged(status int)
	OnUpdates(constructor int64, b []byte)
	OnGeneralError(b []byte)
	OnSessionClosed(res int)
}

// FileDelegate
type FileDelegate interface {
	OnProgressChanged(messageID, processedParts, totalParts int64, percent float64)
	OnCompleted(messageID int64, filePath string)
	OnError(messageID int64, filePath string, err []byte)
}

type ConnInfoDelegate interface {
	SaveConnInfo(connInfo []byte)
}

// RequestDelegate each request should have this callbacks
type RequestDelegate interface {
	OnComplete(b []byte)
	OnTimeout(err error)
}
