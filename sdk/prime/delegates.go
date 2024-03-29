package riversdk

import (
    "github.com/ronaksoft/river-sdk/internal/request"
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
    OnUpdate(action int32, b []byte)
    InitStream(audio, video bool) bool
    InitConnection(connID int32, b []byte) int64
    CloseConnection(connID int32) bool
    GetOfferSDP(connID int32) (out []byte)
    SetOfferGetAnswerSDP(connID int32, req []byte) (out []byte)
    SetAnswerSDP(connID int32, b []byte) bool
    AddIceCandidate(connID int32, b []byte) bool
}

type RequestDelegateFlag = request.DelegateFlag

// Request Flags
// These are exact copies of request.Delegate flags
const (
    RequestServerForced RequestDelegateFlag = 1 << iota
    RequestBlocking
    RequestSkipWaitForNetwork
    RequestSkipFlusher
    RequestRealtime
    RequestBatch
)

type RequestDelegate interface {
    OnComplete(b []byte)
    OnTimeout(err error)
    Flags() RequestDelegateFlag
}
