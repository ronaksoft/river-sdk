package synchronizer

import (
	"fmt"
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/cmd"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/network"
	"git.ronaksoftware.com/ronak/riversdk/repo"
	"git.ronaksoftware.com/ronak/riversdk/repo/dto"

	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/queue"
	"go.uber.org/zap"
)

// SyncConfig
type SyncConfig struct {
	ConnInfo    domain.RiverConfiger
	NetworkCtrl *network.NetworkController
	QueueCtrl   *queue.QueueController
}

// SyncController
type SyncController struct {
	connInfo domain.RiverConfiger
	//sync.Mutex
	networkCtrl          *network.NetworkController
	queue                *queue.QueueController
	onSyncStatusChange   domain.SyncStatusUpdateCallback
	onUpdateMainDelegate domain.OnUpdateMainDelegateHandler
	// InternalChannel
	stopChannel chan bool

	// Internals
	syncStatus         domain.SyncStatus
	lastUpdateReceived time.Time
	updateID           int64
	updateAppliers     map[int64]domain.UpdateApplier
	messageAppliers    map[int64]domain.MessageApplier
	UserID             int64

	// delivered Message
	deliveredMessagesMutex sync.Mutex
	deliveredMessages      map[int64]bool

	updatingDifferenceLock sync.Mutex
	updatingDifference     bool
}

// NewSyncController
func NewSyncController(config SyncConfig) *SyncController {
	ctrl := new(SyncController)
	ctrl.stopChannel = make(chan bool)
	ctrl.connInfo = config.ConnInfo
	ctrl.queue = config.QueueCtrl
	ctrl.networkCtrl = config.NetworkCtrl

	// set default value to synced status
	ctrl.syncStatus = domain.Synced

	ctrl.updateAppliers = map[int64]domain.UpdateApplier{
		msg.C_UpdateNewMessage:        ctrl.updateNewMessage,
		msg.C_UpdateReadHistoryOutbox: ctrl.updateReadHistoryOutbox,
		msg.C_UpdateReadHistoryInbox:  ctrl.updateReadHistoryInbox,
		msg.C_UpdateMessageEdited:     ctrl.updateMessageEdited,
		msg.C_UpdateMessageID:         ctrl.updateMessageID,
	}

	ctrl.messageAppliers = map[int64]domain.MessageApplier{
		msg.C_AuthAuthorization: ctrl.authAuthorization,
		msg.C_ContactsImported:  ctrl.contactsImported,
		msg.C_ContactsMany:      ctrl.contactsMany,
		msg.C_MessagesDialogs:   ctrl.messagesDialogs,
		msg.C_MessagesSent:      ctrl.messageSent,
		msg.C_AuthSentCode:      ctrl.authSentCode,
		msg.C_UsersMany:         ctrl.usersMany,
		msg.C_MessagesMany:      ctrl.messagesMany,
	}

	ctrl.deliveredMessages = make(map[int64]bool, 0)

	return ctrl
}

// Start
func (ctrl *SyncController) Start() {
	log.LOG.Info("SyncController:: Start")

	// Load the latest UpdateID stored in DB
	if v, err := repo.Ctx().System.LoadInt(domain.CN_UPDATE_ID); err != nil {
		err := repo.Ctx().System.SaveInt(domain.CN_UPDATE_ID, 0)
		if err != nil {
			log.LOG.Debug("SyncController::Start()-> SaveInt()",
				zap.String("Error", err.Error()),
			)
		}
		ctrl.updateID = 0
	} else {
		ctrl.updateID = int64(v)
	}

	// Sync with Server
	go ctrl.sync()

	go ctrl.watchDog()
}

// Stop
func (ctrl *SyncController) Stop() {
	ctrl.stopChannel <- true // for watchDog()
}

// SetSyncStatusChangedCallback
func (ctrl *SyncController) SetSyncStatusChangedCallback(h domain.SyncStatusUpdateCallback) {
	ctrl.onSyncStatusChange = h
}

// SetSyncStatusChangedCallback
func (ctrl *SyncController) SetOnUpdateCallback(h domain.OnUpdateMainDelegateHandler) {
	ctrl.onUpdateMainDelegate = h
}

// updateSyncStatus
func (ctrl *SyncController) updateSyncStatus(newStatus domain.SyncStatus) {
	if ctrl.syncStatus == newStatus {
		log.LOG.Info("SyncController::updateSyncStatus() syncStatus not changed")
		return
	}
	switch newStatus {
	case domain.OutOfSync:
		log.LOG.Info("SyncController::updateSyncStatus() OutOfSync")
	case domain.Syncing:
		log.LOG.Info("SyncController::updateSyncStatus() Syncing")
	case domain.Synced:
		log.LOG.Info("SyncController::updateSyncStatus() Synced")
	}
	ctrl.syncStatus = newStatus

	if ctrl.onSyncStatusChange != nil {
		ctrl.onSyncStatusChange(newStatus)
	}
}

// watchDog
// Checks if we have not received any updates since last watch tries to re-sync with server.
func (ctrl *SyncController) watchDog() {
	for {
		select {
		case <-time.After(60 * time.Second):
			// make sure network is connected b4 start getUpdateDifference or snapshotSync
			for ctrl.networkCtrl.Quality() == domain.DISCONNECTED || ctrl.networkCtrl.Quality() == domain.CONNECTING {
				time.Sleep(100 * time.Millisecond)
			}
			if ctrl.syncStatus != domain.Syncing {
				log.LOG.Info("SyncController::watchDog() -> sync() called")
				ctrl.sync()
			}
		case <-ctrl.stopChannel:
			log.LOG.Info("SyncController::watchDog() Stopped")
			return
		}
	}
}

func (ctrl *SyncController) sync() {
	log.LOG.Debug("SyncController::sync()",
		zap.Int64("UpdateID", ctrl.updateID),
		zap.Int64("UserID", ctrl.UserID),
	)
	if ctrl.UserID == 0 {
		return
	}

	// make sure network is connected b4 start getUpdateDifference or snapshotSync
	for ctrl.networkCtrl.Quality() == domain.DISCONNECTED || ctrl.networkCtrl.Quality() == domain.CONNECTING {
		time.Sleep(100 * time.Millisecond)
	}

	var serverUpdateID int64
	var err error
	// try each 100ms until we get serverUpdateID from server
	for {
		serverUpdateID, err = ctrl.getUpdateState()
		if err != nil {
			log.LOG.Debug("SyncController::sync()-> getUpdateState()",
				zap.String("Error", err.Error()),
			)

			time.Sleep(100 * time.Millisecond)
		} else {
			break
		}
	}

	if ctrl.updateID == 0 || (serverUpdateID-ctrl.updateID) > 999 {

		// remove all messages
		err := repo.Ctx().DropAndCreateTable(&dto.Messages{})
		if err != nil {
			log.LOG.Debug("SyncController::sync()-> DropAndCreateTable()",
				zap.String("Error", err.Error()),
			)
		}
		// Get Contacts from the server
		ctrl.getContacts()
		ctrl.updateID = serverUpdateID
		ctrl.getAllDialogs(0, 100)
		err = repo.Ctx().System.SaveInt(domain.CN_UPDATE_ID, int32(ctrl.updateID))
		if err != nil {
			log.LOG.Debug("SyncController::sync()-> SaveInt()",
				zap.String("Error", err.Error()),
			)
		}
	} else if time.Now().Sub(ctrl.lastUpdateReceived).Truncate(time.Second) > 60 {
		// if it is passed over 60 seconds from the last update received it fetches the update
		// difference from the server
		if serverUpdateID > ctrl.updateID+1 {
			ctrl.updateSyncStatus(domain.OutOfSync)
			ctrl.getUpdateDifference(serverUpdateID)
		}
	}
}

// SetUserID
func (ctrl *SyncController) SetUserID(userID int64) {
	ctrl.UserID = userID
	log.LOG.Debug("SyncController::SetUserID()",
		zap.Int64("UserID", userID),
	)
}

// getUpdateState responsibility is to only get server updateID
func (ctrl *SyncController) getUpdateState() (updateID int64, err error) {

	// when network is disconnected no need to enqueue update request in goque
	if ctrl.networkCtrl.Quality() == domain.DISCONNECTED || ctrl.networkCtrl.Quality() == domain.CONNECTING {
		return -1, domain.ErrNoConnection
	}

	req := new(msg.UpdateGetState)
	reqBytes, _ := req.Marshal()
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(1)
	ctrl.queue.ExecuteCommand(
		domain.RandomUint64(),
		msg.C_UpdateGetState,
		reqBytes,
		func() {
			defer waitGroup.Done()
			err = domain.ErrRequestTimeout
		},
		func(m *msg.MessageEnvelope) {
			defer waitGroup.Done()

			switch m.Constructor {
			case msg.C_UpdateState:
				x := new(msg.UpdateState)
				x.Unmarshal(m.Message)
				updateID = x.UpdateID
			case msg.C_Error:
				err = domain.ServerError(m.Message)
				log.LOG.Debug(err.Error())
			}
		},
	)
	waitGroup.Wait()
	return
}

// getUpdateDifference
func (ctrl *SyncController) getUpdateDifference(minUpdateID int64) {

	log.LOG.Debug("SyncController::getUpdateDifference()")
	ctrl.updatingDifferenceLock.Lock()
	if ctrl.updatingDifference {
		ctrl.updatingDifferenceLock.Unlock()
		log.LOG.Debug("SyncController::getUpdateDifference() Exited already updatingDifference")
		return
	}
	ctrl.updatingDifference = true
	ctrl.updatingDifferenceLock.Unlock()

	// if updateID is zero then wait for snapshot sync
	// and when sending requests w8 till its finish
	if ctrl.updateID == 0 {
		ctrl.updatingDifference = false
		log.LOG.Debug("SyncController::getUpdateDifference() Exited UpdateID is zero need snapshot sync")
		return
	}

	for minUpdateID > ctrl.updateID {
		limit := minUpdateID - ctrl.updateID
		limit++
		if limit > 100 {
			limit = 100
		}
		log.LOG.Debug("SyncController::getUpdateDifference() Entered loop",
			zap.Int64("limit", limit),
			zap.Int64("updateID", ctrl.updateID),
			zap.Int64("minUpdateID", minUpdateID),
		)

		ctrl.updateSyncStatus(domain.Syncing)
		req := new(msg.UpdateGetDifference)
		req.Limit = int32(limit)
		req.From = ctrl.updateID
		reqBytes, _ := req.Marshal()
		ctrl.queue.ExecuteRealtimeCommand(
			domain.RandomUint64(),
			msg.C_UpdateGetDifference,
			reqBytes,
			func() {
				log.LOG.Debug("SyncController::getUpdateDifference() -> ExecuteRealtimeCommand() Timeout")
			},
			func(m *msg.MessageEnvelope) {
				ctrl.onGetDiffrenceSucceed(m)
				log.LOG.Debug("SyncController::getUpdateDifference() -> ExecuteRealtimeCommand() Success")
			},
			true,
		)
		log.LOG.Debug("SyncController::getUpdateDifference() Loop next")
	}
	log.LOG.Debug("SyncController::getUpdateDifference() Loop Finished")
	ctrl.updatingDifference = false
	ctrl.updateSyncStatus(domain.Synced)
}

func (ctrl *SyncController) onGetDiffrenceSucceed(m *msg.MessageEnvelope) {
	switch m.Constructor {
	case msg.C_UpdateDifference:
		x := new(msg.UpdateDifference)
		err := x.Unmarshal(m.Message)
		if err != nil {
			log.LOG.Debug("SyncController::onGetDiffrenceSucceed()-> Unmarshal()",
				zap.String("Error", err.Error()),
			)
			return
		}
		updContainer := new(msg.UpdateContainer)
		updContainer.Updates = make([]*msg.UpdateEnvelope, 0)
		updContainer.Users = x.Users
		updContainer.MaxUpdateID = x.MaxUpdateID
		updContainer.MinUpdateID = x.MinUpdateID

		log.LOG.Warn("SyncController::onGetDiffrenceSucceed()",
			zap.Int64("UpdateID", ctrl.updateID),
			zap.Int64("MaxUpdateID", x.MaxUpdateID),
			zap.Int64("MinUpdateID", x.MinUpdateID),
		)

		// take out updateMessageID
		for _, update := range x.Updates {
			if update.Constructor == msg.C_UpdateMessageID {
				ctrl.updateMessageID(update)
			}
		}

		for _, update := range x.Updates {

			// we allready processed this update type
			if update.Constructor == msg.C_UpdateMessageID {
				continue
			}

			log.LOG.Debug("SyncController::onGetDiffrenceSucceed() loop",
				zap.String("Constructor", msg.ConstructorNames[update.Constructor]),
			)

			if applier, ok := ctrl.updateAppliers[update.Constructor]; ok {

				// TODO: hotfix added this cuz sync works awkwardly
				passToExternalHandler := applier(update)
				// check updateID to ensure to do not pass old updates to UI.
				// TODO : server still sends typing updates
				if passToExternalHandler && update.UpdateID > ctrl.updateID && update.Constructor != msg.C_UpdateUserTyping {
					updContainer.Updates = append(updContainer.Updates, update)
				}
			}
		}
		updContainer.Length = int32(len(updContainer.Updates))

		// wrapped to UpdateContainer
		buff, _ := updContainer.Marshal()
		cmd.GetUIExecuter().Exec(func() { ctrl.onUpdateMainDelegate(msg.C_UpdateContainer, buff) })

		// update last updateID
		if ctrl.updateID < x.MaxUpdateID {
			ctrl.updateID = x.MaxUpdateID

			// Save UpdateID to DB
			err := repo.Ctx().System.SaveInt(domain.CN_UPDATE_ID, int32(ctrl.updateID))
			if err != nil {
				log.LOG.Debug("SyncController::onGetDiffrenceSucceed()-> SaveInt()",
					zap.String("Error", err.Error()),
				)
			}
		}
	case msg.C_Error:
		log.LOG.Debug("SyncController::onGetDiffrenceSucceed()-> C_Error",
			zap.String("Error", domain.ServerError(m.Message).Error()),
		)
		// TODO:: Handle error
	}
}

// getContacts
func (ctrl *SyncController) getContacts() {
	req := new(msg.ContactsGet)
	reqBytes, _ := req.Marshal()
	ctrl.queue.ExecuteCommand(
		domain.RandomUint64(),
		msg.C_ContactsGet,
		reqBytes,
		nil,
		func(m *msg.MessageEnvelope) {
			// SyncController applier will take care of this
		},
	)
}

// getAllDialogs
func (ctrl *SyncController) getAllDialogs(offset int32, limit int32) {
	log.LOG.Info("SyncController::getAllDialogs()")
	req := new(msg.MessagesGetDialogs)
	req.Limit = limit
	req.Offset = offset
	reqBytes, _ := req.Marshal()
	ctrl.queue.ExecuteCommand(
		domain.RandomUint64(),
		msg.C_MessagesGetDialogs,
		reqBytes,
		func() {
			log.LOG.Info("SyncController::getAllDialogs() -> onTimeoutback() retry to getAllDialogs()")
			ctrl.getAllDialogs(offset, limit)
		},
		func(m *msg.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_MessagesDialogs:
				x := new(msg.MessagesDialogs)
				err := x.Unmarshal(m.Message)
				if err != nil {
					log.LOG.Info("SyncController::getAllDialogs() -> onSuccessCallback() -> Unmarshal() ",
						zap.String("Error", err.Error()),
					)
					return
				}
				log.LOG.Debug("SyncController::getAllDialogs() -> onSuccessCallback() -> MessagesDialogs",
					zap.Int("DialogsLength", len(x.Dialogs)),
					zap.Int32("Offset", offset),
					zap.Int32("Total", x.Count),
				)
				mMessages := make(map[int64]*msg.UserMessage)
				for _, message := range x.Messages {
					err := repo.Ctx().Messages.SaveMessage(message)
					if err != nil {
						log.LOG.Info("SyncController::getAllDialogs() -> onSuccessCallback() -> SaveMessage() ",
							zap.String("Error", err.Error()),
						)
					}

					mMessages[message.ID] = message
				}

				for _, dialog := range x.Dialogs {
					topMessage, _ := mMessages[dialog.TopMessageID]
					if topMessage == nil {
						log.LOG.Info("SyncController::getAllDialogs() -> onSuccessCallback() -> dialog TopMessage is null ",
							zap.Int64("MessageID", dialog.TopMessageID),
						)
						continue
					}
					err := repo.Ctx().Dialogs.SaveDialog(dialog, topMessage.CreatedOn)
					if err != nil {
						log.LOG.Info("SyncController::getAllDialogs() -> onSuccessCallback() -> SaveDialog() ",
							zap.String("Error", err.Error()),
							zap.String("Dialog", fmt.Sprintf("%v", dialog)),
						)
					}
				}
				for _, user := range x.Users {
					err := repo.Ctx().Users.SaveUser(user)
					if err != nil {
						log.LOG.Info("SyncController::getAllDialogs() -> onSuccessCallback() -> SaveUser() ",
							zap.String("Error", err.Error()),
							zap.String("User", fmt.Sprintf("%v", user)),
						)
					}
				}
				if x.Count > offset+limit {
					log.LOG.Info("SyncController::getAllDialogs() -> onSuccessCallback() retry to getAllDialogs()",
						zap.Int32("x.Count", x.Count),
						zap.Int32("offset+limit", offset+limit),
					)
					ctrl.getAllDialogs(offset+limit, limit)
				}
			case msg.C_Error:
				log.LOG.Debug("SyncController::onSuccessCallback()-> C_Error",
					zap.String("Error", domain.ServerError(m.Message).Error()),
				)
			}
		},
	)

}

func (ctrl *SyncController) addToDeliveredMessageList(id int64) {
	ctrl.deliveredMessagesMutex.Lock()
	ctrl.deliveredMessages[id] = true
	ctrl.deliveredMessagesMutex.Unlock()
}

func (ctrl *SyncController) isDeliveredMessage(id int64) bool {
	ctrl.deliveredMessagesMutex.Lock()
	var ok bool
	if _, ok = ctrl.deliveredMessages[id]; ok {
		// cuz server sends duplicated updates again and again :|
		// delete(ctrl.deliveredMessages, id)
	}
	ctrl.deliveredMessagesMutex.Unlock()

	return ok
}

// Status displays SyncStatus
func (ctrl *SyncController) Status() domain.SyncStatus {
	return ctrl.syncStatus
}

// MessageHandler call appliers-> repository and sync data
func (ctrl *SyncController) MessageHandler(messages []*msg.MessageEnvelope) {
	for _, m := range messages {
		log.LOG.Debug("SyncController::MessageHandler() Received",
			zap.String("Constructor", msg.ConstructorNames[m.Constructor]),
		)
		if applier, ok := ctrl.messageAppliers[m.Constructor]; ok {
			applier(m)
			log.LOG.Debug("SyncController::MessageHandler() Message Applied",
				zap.String("Constructor", msg.ConstructorNames[m.Constructor]),
			)
		}
	}

}

// UpdateHandler
func (ctrl *SyncController) UpdateHandler(u *msg.UpdateContainer) {

	if ctrl.syncStatus != domain.Synced {
		log.LOG.Debug("SyncController::UpdateHandler() Ignore updates while syncing")
		return
	}
	log.LOG.Debug("SyncController::UpdateHandler() Called")
	ctrl.lastUpdateReceived = time.Now()
	if u.MinUpdateID != 0 {
		// Check if we are out of sync with server, if yes, then get the difference and
		// try to sync with server again
		if ctrl.updateID < u.MinUpdateID-1 {
			ctrl.updateSyncStatus(domain.OutOfSync)
			ctrl.getUpdateDifference(u.MinUpdateID)
		}

		if ctrl.updateID < u.MaxUpdateID {
			ctrl.updateID = u.MaxUpdateID
			err := repo.Ctx().System.SaveInt(domain.CN_UPDATE_ID, int32(ctrl.updateID))
			if err != nil {
				log.LOG.Debug("SyncController::UpdateHandler() -> SaveInt()",
					zap.String("Error", err.Error()),
				)
			}
		}
	}

	udpContainer := new(msg.UpdateContainer)
	udpContainer.Updates = make([]*msg.UpdateEnvelope, 0)
	udpContainer.MaxUpdateID = u.MaxUpdateID
	udpContainer.MinUpdateID = u.MinUpdateID

	// take out updateMessageID
	for _, update := range u.Updates {
		if update.Constructor == msg.C_UpdateMessageID {
			ctrl.updateMessageID(update)
		}
	}

	for _, update := range u.Updates {

		// we allready processed this updte type
		if update.Constructor == msg.C_UpdateMessageID {
			continue
		}

		log.LOG.Debug("SyncController::UpdateHandler() Update Received",
			zap.String("Constructor", msg.ConstructorNames[update.Constructor]),
		)

		// this will be handled by update applier
		// // save MessageID and discard update on updateNewMessage cuz its sent by yourself
		// if update.Constructor == msg.C_UpdateMessageID {
		// 	m := new(msg.UpdateMessageID)
		// 	m.Unmarshal(update.Update)
		// 	ctrl.addToDeliveredMessageList(m.MessageID)
		// 	continue
		// }

		passToExternalHandler := true
		if applier, ok := ctrl.updateAppliers[update.Constructor]; ok {

			passToExternalHandler = applier(update)
			log.LOG.Debug("SyncController::UpdateHandler() Update Applied",
				zap.Int64("UPDATE_ID", update.UpdateID),
				zap.String("Constructor", msg.ConstructorNames[update.Constructor]),
			)
		}

		if passToExternalHandler {
			udpContainer.Updates = append(udpContainer.Updates, update)

		} else {
			log.LOG.Debug("SyncController::UpdateHandler() Do not pass update to external handler",
				zap.Int64("UPDATE_ID", update.UpdateID),
				zap.String("Constructor", msg.ConstructorNames[update.Constructor]),
			)
		}

	}
	udpContainer.Length = int32(len(udpContainer.Updates))

	// call external handler
	if ctrl.onUpdateMainDelegate != nil {

		// wrapped to UpdateContainer
		buff, _ := udpContainer.Marshal()

		// pass all updates to UI
		cmd.GetUIExecuter().Exec(func() {
			ctrl.onUpdateMainDelegate(msg.C_UpdateContainer, buff)
		})
	}

}
