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

var (
	dlg MainDelegate
)

func getMainDelegate() MainDelegate {
	if dlg == nil {
		panic("main delegates not initialized")
	}
	return dlg
}

func setMainDelegate(d MainDelegate) {
	dlg = d
}
