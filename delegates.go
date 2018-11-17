package riversdk

type MainDelegate interface {
	OnNetworkStatusChanged(status int)
	OnSyncStatusChanged(status int)
	OnDeferredRequests(requestID int64, b []byte)
	OnUpdates(constructor int64, b []byte)
	OnAuthKeyCreated(int64)
	OnGeneralError(b []byte)
	OnSessionClosed(res int)
}

type RequestDelegate interface {
	OnComplete(b []byte)
	OnTimeout(err error)
}

type LoggerDelegate interface {
	Log(logLevel int, msg string)
}
