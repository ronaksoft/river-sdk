package riversdk

// MainDelegate extenal (UI) handler will listen to this function to receive data from SDK
type MainDelegate interface {
	OnNetworkStatusChanged(status int)
	OnSyncStatusChanged(status int)
	OnDeferredRequests(requestID int64, b []byte)
	OnUpdates(constructor int64, b []byte)
	OnAuthKeyCreated(int64)
	OnGeneralError(b []byte)
	OnSessionClosed(res int)

	OnDownloadProgressChanged(messageID, processedParts, totalParts int64, percent float64)
	OnUploadProgressChanged(messageID, processedParts, totalParts int64, percent float64)
	OnDownloadCompleted(messageID int64, filePath string)
	OnUploadCompleted(messageID int64, filePath string)
	OnDownloadError(messageID, requestID int64, filePath string, err []byte)
	OnUploadError(messageID, requestID int64, filePath string, err []byte)
}

// RequestDelegate each request shoudl have this callbacks
type RequestDelegate interface {
	OnComplete(b []byte)
	OnTimeout(err error)
}

// LoggerDelegate callback to attack logs to extenal (UI) handler
type LoggerDelegate interface {
	Log(logLevel int, msg string)
}
