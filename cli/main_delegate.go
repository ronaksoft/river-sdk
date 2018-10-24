package main

import (
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"go.uber.org/zap"
)

type MainDelegate struct{}

func (d *MainDelegate) OnUpdates(constructor int64, b []byte) {

	switch constructor {
	case msg.C_UpdateContainer:
		updateContainer := new(msg.UpdateContainer)
		err := updateContainer.Unmarshal(b)
		if err != nil {
			_Log.Debug(err.Error())
			return
		}
		_MAGNETA("UpdateContainer:: %d --> %d", updateContainer.MinUpdateID, updateContainer.MaxUpdateID)
		for _, update := range updateContainer.Updates {
			UpdatePrinter(update)
		}
	case msg.C_ClientPendingMessageDelivery:
		//wrapping it to update envelop to pass UpdatePrinter
		udp := new(msg.UpdateEnvelope)
		udp.Constructor = constructor
		udp.Update = b

		UpdatePrinter(udp)
	}

}

func (d *MainDelegate) OnDeferredRequests(requestID int64, b []byte) {
	envelope := new(msg.MessageEnvelope)
	envelope.Unmarshal(b)
	_Shell.Println(_GREEN("RequestID: %d", requestID))
	MessagePrinter(envelope)
}

func (d *MainDelegate) OnNetworkStatusChanged(quality int) {
	status := []string{
		"Disconnected", "Connecting", "Week", "Slow", "Fast",
	}
	_Log.Info("Network Status Changed",
		zap.String("Status", status[quality]))
}

func (d *MainDelegate) OnSyncStatusChanged(newStatus int) {
	status := []string{
		"Out of Sync", "Syncing", "Synced",
	}
	_Log.Info("Network Status Changed",
		zap.String("Status", status[newStatus]))
}

func (d *MainDelegate) OnAuthKeyCreated(authID int64) {
	_Log.Info("Auth Key Created",
		zap.Int64("AuthID", authID),
	)
}

func (d *MainDelegate) OnGeneralError(b []byte) {
	e := new(msg.Error)
	e.Unmarshal(b)

	_Log.Info("OnGeneralError",
		zap.String("Code", e.Code),
		zap.String("Items", e.Items),
	)
}

func (d *MainDelegate) OnSessionClosed(res int) {
	_Log.Info("OnSessionClosed",
		zap.Int("Res", res),
	)
}
