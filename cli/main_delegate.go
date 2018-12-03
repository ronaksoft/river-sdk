package main

import (
	"git.ronaksoftware.com/ronak/riversdk/msg"
)

type MainDelegate struct{}

func (d *MainDelegate) OnUpdates(constructor int64, b []byte) {

	_Shell.Println(_RED("OnUpdates() Constructor: %v", msg.ConstructorNames[constructor]))

	switch constructor {
	case msg.C_UpdateContainer:
		updateContainer := new(msg.UpdateContainer)
		err := updateContainer.Unmarshal(b)
		if err != nil {
			_Log.Debug(err.Error())
			return
		}
		_Shell.Println(_MAGNETA("OnUpdates() :: UpdateContainer:: %d --> %d", updateContainer.MinUpdateID, updateContainer.MaxUpdateID))
		for _, update := range updateContainer.Updates {
			_Shell.Println(_MAGNETA("OnUpdates() :: Loop Update Constructor :: %v", msg.ConstructorNames[update.Constructor]))
			UpdatePrinter(update)
		}
	case msg.C_ClientUpdatePendingMessageDelivery:
		//wrapping it to update envelop to pass UpdatePrinter
		udp := new(msg.UpdateEnvelope)
		udp.Constructor = constructor
		udp.Update = b

		UpdatePrinter(udp)
	case msg.C_UpdateEnvelope:
		update := new(msg.UpdateEnvelope)
		err := update.Unmarshal(b)
		if err != nil {
			_Log.Debug(err.Error())
			return
		} else {
			_Shell.Println(_MAGNETA("OnUpdates() :: Update Constructor :: %v", msg.ConstructorNames[update.Constructor]))
			UpdatePrinter(update)
		}
	}

}

func (d *MainDelegate) OnDeferredRequests(requestID int64, b []byte) {
	envelope := new(msg.MessageEnvelope)
	envelope.Unmarshal(b)
	_Shell.Println(_RED("OnDeferredRequests() RequestID: %d", requestID))
	MessagePrinter(envelope)
}

func (d *MainDelegate) OnNetworkStatusChanged(quality int) {
	status := []string{
		"Disconnected", "Connecting", "Week", "Slow", "Fast",
	}
	_Shell.Println(_RED("Network Status Changed: Status = %v", status[quality]))
}

func (d *MainDelegate) OnSyncStatusChanged(newStatus int) {
	status := []string{
		"Out of Sync", "Syncing", "Synced",
	}
	_Shell.Println(_RED("Sync Status Changed: Status = %v", status[newStatus]))
}

func (d *MainDelegate) OnAuthKeyCreated(authID int64) {

	_Shell.Println(_RED("Auth Key Created: AuthID = %v", authID))
}

func (d *MainDelegate) OnGeneralError(b []byte) {
	e := new(msg.Error)
	e.Unmarshal(b)

	_Shell.Println(_RED("OnGeneralError: {Code = %v , Items = %v }", e.Code, e.Items))
}

func (d *MainDelegate) OnSessionClosed(res int) {
	_Shell.Println(_RED("OnSessionClosed : Res = %v", res))
}
