package riversdk

// MainDelegate external (UI) handler will listen to this function to receive data from SDK
type MainDelegate interface {
	OnNetworkStatusChanged(status int)
	OnSyncStatusChanged(status int)
	OnUpdates(constructor int64, b []byte)
	OnGeneralError(b []byte)
	OnSessionClosed(res int)
	ShowLoggerAlert()
	AddLog(text string)
}

// FileDelegate
type FileDelegate interface {
	OnProgressChanged(reqID string, clusterID int32, fileID, accessHash, percent int64)
	OnCompleted(reqID string, clusterID int32, fileID, accessHash int64, filePath string)
	OnCancel(reqID string, clusterID int32, fileID, accessHash int64, hasError bool)
}

type ConnInfoDelegate interface {
	SaveConnInfo(connInfo []byte)
	Get(key string) string
	Set(key, value string)
}

// RequestDelegate each request should have this callbacks
type RequestDelegate interface {
	OnComplete(b []byte)
	OnTimeout(err error)
	Flags() int32
}

// Request Flags
const (
	RequestServerForced int32 = 1 << iota
	RequestBlocking
	RequestDontWaitForNetwork
)
