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
	AppUpdate(version string, updateAvailable, force bool)
	DataSynced(dialogs, contacts, gifs bool)
}

type FileDelegate interface {
	OnProgressChanged(reqID string, clusterID int32, fileID, accessHash, percent int64, peerID int64)
	OnCompleted(reqID string, clusterID int32, fileID, accessHash int64, filePath string, peerID int64)
	OnCancel(reqID string, clusterID int32, fileID, accessHash int64, hasError bool, peerID int64)
}

type ConnInfoDelegate interface {
	SaveConnInfo(connInfo []byte)
	Get(key string) string
	Set(key, value string)
}

type CallDelegate interface {
	OnUpdate(constructor int64, b []byte)
	InitStream(audio, video bool) (err error)
	InitConnection(connId int32, b []byte) (id int64, err error)
	CloseConnection(connId int32, all bool) (err error)
	GetAnswerSDP(connId int32, req []byte) (res []byte, err error)
	GetOfferSDP(connId int32) (res []byte, err error)
	SetAnswerSDP(connId int32, b []byte) (err error)
	AddIceCandidate(connId int32, b []byte) (err error)
}

// Request Flags
const (
	RequestServerForced int32 = 1 << iota
	RequestBlocking
	RequestDontWaitForNetwork
	RequestTeamForce
)

type RequestDelegate interface {
	OnComplete(b []byte)
	OnTimeout(err error)
	Flags() int32
}
