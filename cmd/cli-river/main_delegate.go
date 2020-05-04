package main

import (
	"git.ronaksoftware.com/river/msg/chat"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
)



type ConnInfoDelegates struct{
	filePath 	string
}

func (c *ConnInfoDelegates) Get(key string) string {
	panic("implement me")
}

func (c *ConnInfoDelegates) Set(key, value string) {
	panic("implement me")
}

func (c *ConnInfoDelegates) SaveConnInfo(connInfo []byte) {
	_ = os.MkdirAll("./_connection", os.ModePerm)
	err := ioutil.WriteFile(c.filePath, connInfo, 0666)
	if err != nil {
		_Log.Error(err.Error())
	}
}

type MainDelegate struct{}

func (d *MainDelegate) OnSearchComplete(b []byte) {
	_Log.Info("OnSearchComplete")
	result := new(msg.ClientSearchResult)
	err := result.Unmarshal(b)
	if err != nil {
		_Log.Warn("error Unmarshal", zap.String("", err.Error()))
		return
	}
	_Shell.Println("OnSearchComplete::Messages", result.Messages)
	_Shell.Println("OnSearchComplete::Groups", result.Groups)
	_Shell.Println("OnSearchComplete::MatchedGroups", result.MatchedGroups)
	_Shell.Println("OnSearchComplete::MatchedUsers", result.MatchedUsers)
}

func (d *MainDelegate) OnUpdates(constructor int64, b []byte) {
	_Log.Info("Update received", zap.String("C", msg.ConstructorNames[constructor]))
	switch constructor {
	case msg.C_UpdateContainer:
		updateContainer := new(msg.UpdateContainer)
		err := updateContainer.Unmarshal(b)
		if err != nil {
			_Log.Error("Failed to unmarshal", zap.Error(err))
			return
		}
		_Log.Info("Processing UpdateContainer", zap.Int64("MinID", updateContainer.MinUpdateID), zap.Int64("MaxID", updateContainer.MaxUpdateID))
		for _, update := range updateContainer.Updates {
			_Log.Info("Processing Update", zap.Int64("UpdateID", update.UpdateID), zap.String("C", msg.ConstructorNames[update.Constructor]))
			UpdatePrinter(update)
		}
	case msg.C_ClientUpdatePendingMessageDelivery:
		// wrapping it to update envelop to pass UpdatePrinter
		udp := new(msg.UpdateEnvelope)
		udp.Constructor = constructor
		udp.Update = b
		_Log.Info("Processing ClientUpdatePendingMessageDelivery")
		UpdatePrinter(udp)
	case msg.C_UpdateEnvelope:
		update := new(msg.UpdateEnvelope)
		err := update.Unmarshal(b)
		if err != nil {
			_Log.Error("Failed to unmarshal", zap.Error(err))
			return
		} else {
			_Log.Info("Processing UpdateEnvelop", zap.Int64("UpdateID", update.UpdateID), zap.String("C", msg.ConstructorNames[update.Constructor]))
			UpdatePrinter(update)
		}
	}

}

func (d *MainDelegate) OnDeferredRequests(requestID int64, b []byte) {
	envelope := new(msg.MessageEnvelope)
	envelope.Unmarshal(b)
	_Log.Info("Deferred Request received",
		zap.Uint64("ReqID", envelope.RequestID),
		zap.String("C", msg.ConstructorNames[envelope.Constructor]),
	)
	MessagePrinter(envelope)
}

func (d *MainDelegate) OnNetworkStatusChanged(quality int) {
	state := domain.NetworkStatus(quality)
	_Log.Info("Network status changed", zap.String("Status", state.ToString()))
}

func (d *MainDelegate) OnSyncStatusChanged(newStatus int) {
	state := domain.SyncStatus(newStatus)
	_Log.Info("Sync status changed", zap.String("Status", state.ToString()))
}

func (d *MainDelegate) OnAuthKeyCreated(authID int64) {
	_Log.Info("Auth Key Created", zap.Int64("AuthID", authID))
}

func (d *MainDelegate) OnGeneralError(b []byte) {
	e := new(msg.Error)
	e.Unmarshal(b)
	_Log.Error("Received general error", zap.String("Code", e.Code), zap.String("Items", e.Items))
}

func (d *MainDelegate) OnSessionClosed(res int) {
	_Log.Info("Session Closed", zap.Int("Res", res))
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
	_Log.Info("upload progress changed",
		zap.Int64("Progress", percent),
		zap.Int64("PeerID", peerID),
	)
}

func (d *FileDelegate) OnCompleted(reqID string, clusterID int32, fileID, accessHash int64, filePath string, peerID int64) {
	_Log.Info("On upload Completed",
		zap.String("ReqID", reqID),
		zap.String("FilePath", filePath),
		zap.Int64("PeerID", peerID),
	)
}

func (d *FileDelegate) OnCancel(reqID string, clusterID int32, fileID, accessHash int64, hasError bool, peerID int64) {
	_Log.Error("CancelCB",
		zap.String("ReqID", reqID),
		zap.Int64("PeerID", peerID),
	)

}

func (d *FileDelegate) OnDownloadProgressChanged(messageID, processedParts, totalParts int64, percent float64) {
	_Log.Info("upload progress changed", zap.Float64("Progress", percent))
}

func (d *FileDelegate) OnDownloadCompleted(messageID int64, filePath string) {
	_Log.Info("On upload Completed", zap.Int64("MsgID", messageID), zap.String("FilePath", filePath))
}

func (d *FileDelegate) OnDownloadError(messageID, requestID int64, filePath string, err []byte) {
	x := new(msg.Error)
	x.Unmarshal(err)

	_Log.Error("OnError",
		zap.String("Code", x.Code),
		zap.String("Item", x.Items),
		zap.Int64("MsgID", messageID),
		zap.Int64("ReqID", requestID),
		zap.String("FilePath", filePath),
	)

}
