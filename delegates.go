package riversdk

// MainDelegate external (UI) handler will listen to this function to receive data from SDK
type MainDelegate interface {
	OnNetworkStatusChanged(status int)
	OnSyncStatusChanged(status int)
	OnUpdates(constructor int64, b []byte)
	OnGeneralError(b []byte)
	OnSessionClosed(res int)

	OnDownloadProgressChanged(messageID, processedParts, totalParts int64, percent float64)
	OnUploadProgressChanged(messageID, processedParts, totalParts int64, percent float64)
	OnDownloadCompleted(messageID int64, filePath string)
	OnUploadCompleted(messageID int64, filePath string)
	OnDownloadError(messageID, requestID int64, filePath string, err []byte)
	OnUploadError(messageID, requestID int64, filePath string, err []byte)
}

type ConnInfoDelegate interface {
	LoadConnInfo() (connInfo []byte, err error)
	SaveConnInfo(connInfo []byte)
}

// RequestDelegate each request should have this callbacks
type RequestDelegate interface {
	OnComplete(b []byte)
	OnTimeout(err error)
}

// LoggerDelegate callback to attack logs to external (UI) handler
type LoggerDelegate interface {
	Log(logLevel int, msg string)
}
