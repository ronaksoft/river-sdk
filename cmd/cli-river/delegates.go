package main

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"github.com/ronaksoft/rony"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
)

type ConnInfoDelegates struct {
	dbPath   string
	filePath string
}

func (c *ConnInfoDelegates) Get(key string) string {
	panic("implement me")
}

func (c *ConnInfoDelegates) Set(key, value string) {
	panic("implement me")
}

func (c *ConnInfoDelegates) SaveConnInfo(connInfo []byte) {
	_ = os.MkdirAll(c.dbPath, os.ModePerm)
	err := ioutil.WriteFile(c.filePath, connInfo, 0666)
	if err != nil {
		_Shell.Println(err)
	}
}

type MainDelegate struct{}

func (d *MainDelegate) OnSearchComplete(b []byte) {
	result := new(msg.ClientSearchResult)
	err := result.Unmarshal(b)
	if err != nil {
		_Shell.Println("Error On OnSearchComplete:", err.Error())
		return
	}
	_Shell.Println("OnSearchComplete::Messages", result.Messages)
	_Shell.Println("OnSearchComplete::Groups", result.Groups)
	_Shell.Println("OnSearchComplete::MatchedGroups", result.MatchedGroups)
	_Shell.Println("OnSearchComplete::MatchedUsers", result.MatchedUsers)
}

func (d *MainDelegate) OnUpdates(constructor int64, b []byte) {
	switch constructor {
	case msg.C_UpdateContainer:
		updateContainer := new(msg.UpdateContainer)
		err := updateContainer.Unmarshal(b)
		if err != nil {
			_Shell.Println("Failed To Unmarshal UpdateContainer:", err)
			return
		}
		// _Shell.Println("Processing UpdateContainer:", updateContainer.MinUpdateID, updateContainer.MaxUpdateID)
		for _, update := range updateContainer.Updates {
			// _Shell.Println("Processing Update", update.UpdateID, msg.ConstructorNames[update.Constructor])
			UpdatePrinter(update)
		}
	case msg.C_ClientUpdatePendingMessageDelivery:
		// wrapping it to update envelop to pass UpdatePrinter
		udp := new(msg.UpdateEnvelope)
		udp.Constructor = constructor
		udp.Update = b
		// _Shell.Println("Processing ClientUpdatePendingMessageDelivery")
		UpdatePrinter(udp)
	case msg.C_UpdateEnvelope:
		update := new(msg.UpdateEnvelope)
		err := update.Unmarshal(b)
		if err != nil {
			_Shell.Println("Error On Unmarshal UpdateEnvelope:", err)
			return
		} else {
			// _Shell.Println("Processing UpdateEnvelop", update.UpdateID, msg.ConstructorNames[update.Constructor])
			UpdatePrinter(update)
		}
	}

}

func (d *MainDelegate) OnDeferredRequests(requestID int64, b []byte) {
	envelope := new(rony.MessageEnvelope)
	envelope.Unmarshal(b)
	_Shell.Println("Deferred Request received",
		zap.Uint64("ReqID", envelope.RequestID),
		zap.String("C", msg.ConstructorNames[envelope.Constructor]),
	)
	// MessagePrinter(envelope)
}

func (d *MainDelegate) OnNetworkStatusChanged(quality int) {
	// state := domain.NetworkStatus(quality)
	// _Shell.Println("Network status changed:", state.ToString())
}

func (d *MainDelegate) OnSyncStatusChanged(newStatus int) {
	// state := domain.SyncStatus(newStatus)
	// _Shell.Println("Sync status changed:", state.ToString())
}

func (d *MainDelegate) OnAuthKeyCreated(authID int64) {
	_Shell.Println("Auth Key Created", zap.Int64("AuthID", authID))
}

func (d *MainDelegate) OnGeneralError(b []byte) {
	e := new(rony.Error)
	e.Unmarshal(b)
	_Shell.Println("Received general error", zap.String("Code", e.Code), zap.String("Items", e.Items))
}

func (d *MainDelegate) OnSessionClosed(res int) {
	_Shell.Println("Session Closed:", res)
}

func (d *MainDelegate) ShowLoggerAlert() {}

func (d *MainDelegate) AddLog(txt string) {}

func (d *MainDelegate) AppUpdate(version string, available, force bool) {}

type PrintDelegate struct{}

func (d *PrintDelegate) Log(logLevel int, msg string) {
	switch logLevel {
	case int(zap.DebugLevel):
		_Shell.Println("DBG : \t", msg)
	case int(zap.WarnLevel):
		_Shell.Println(yellow("WRN : \t %s", msg))
	case int(zap.InfoLevel):
		_Shell.Println(green("INF : \t %s", msg))
	case int(zap.ErrorLevel):
		_Shell.Println(red("ERR : \t %s", msg))
	case int(zap.FatalLevel):
		_Shell.Println(red("FTL : \t %s", msg))
	default:
		_Shell.Println(blue("MSG : \t %s", msg))
	}
}

type FileDelegate struct{}

func (d *FileDelegate) OnProgressChanged(reqID string, clusterID int32, fileID, accessHash int64, percent int64, peerID int64) {
	// _Shell.Println("File Progress Changed", reqID, fileID, percent)
}

func (d *FileDelegate) OnCompleted(reqID string, clusterID int32, fileID, accessHash int64, filePath string, peerID int64) {
	// _Shell.Println("File Progress Completed", reqID, filePath)
}

func (d *FileDelegate) OnCancel(reqID string, clusterID int32, fileID, accessHash int64, hasError bool, peerID int64) {
	// _Shell.Println("File Progress Canceled", reqID, hasError)
}

type RequestDelegate struct {
	RequestID int64
	Envelope  rony.MessageEnvelope
	FlagsVal  int32
}

func (d *RequestDelegate) OnComplete(b []byte) {
	err := d.Envelope.Unmarshal(b)
	if err != nil {
		_Shell.Println("Error On OnComplete:", err)
		return
	}
	_Shell.Println("Request Completed:", d.RequestID, msg.ConstructorNames[d.Envelope.Constructor])
	MessagePrinter(&d.Envelope)
	return
}

func (d *RequestDelegate) OnTimeout(err error) {
	_Shell.Println("Request TimedOut:", d.RequestID, err)
}

func (d *RequestDelegate) Flags() int32 {
	return d.FlagsVal
}

type CustomRequestDelegate struct {
	RequestID      int64
	OnCompleteFunc func(b []byte)
	OnTimeoutFunc  func(err error)
	FlagsFunc      func() int32
}

func (c CustomRequestDelegate) OnComplete(b []byte) {
	c.OnCompleteFunc(b)
}

func (c CustomRequestDelegate) OnTimeout(err error) {
	c.OnTimeoutFunc(err)
}

func (c CustomRequestDelegate) Flags() int32 {
	return c.FlagsFunc()
}

func NewCustomDelegate() *CustomRequestDelegate {
	c := &CustomRequestDelegate{}
	d := &RequestDelegate{}
	c.OnCompleteFunc = d.OnComplete
	c.OnTimeoutFunc = d.OnTimeout
	c.FlagsFunc = d.Flags
	return c
}
