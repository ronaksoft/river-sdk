package riversdk

type MainDelegate interface {
	OnNetworkStatusChanged(status int)
	OnSyncStatusChanged(status int)
	OnDeferredRequests(requestID int64, b []byte)
	OnUpdates(constructor int64, b []byte)
	OnAuthKeyCreated(int64)
	OnGeneralError(b []byte)
	OnSessionClosed(res int)

	OnDownloadProgressChanged(messageID, position, totalSize int64, percent float64)
	OnUploadProgressChanged(messageID, position, totalSize int64, percent float64)
	OnDownloadCompleted(messageID int64, filePath string)
	OnUploadCompleted(messageID int64, filePath string)
	OnDownloadError(messageID, requestID int64, filePath string, err []byte)
	OnUploadError(messageID, requestID int64, filePath string, err []byte)
}

type RequestDelegate interface {
	OnComplete(b []byte)
	OnTimeout(err error)
}

type LoggerDelegate interface {
	Log(logLevel int, msg string)
}
