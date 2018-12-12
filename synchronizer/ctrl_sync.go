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

	isUpdatingDifferenceLock sync.Mutex
	isUpdatingDifference     bool

	isSyncingLock sync.Mutex
	isSyncing     bool
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
		msg.C_UpdateNotifySettings:    ctrl.updateNotifySettings,
		msg.C_UpdateUsername:          ctrl.updateUsername,

		msg.C_UpdateMessagesDeleted: ctrl.updateMessagesDeleted,
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
		msg.C_GroupFull:         ctrl.groupFull,
	}

	ctrl.deliveredMessages = make(map[int64]bool, 0)

	return ctrl
}

// Start
func (ctrl *SyncController) Start() {
	log.LOG_Info("SyncController:: Start")

	// Load the latest UpdateID stored in DB
	if v, err := repo.Ctx().System.LoadInt(domain.CN_UPDATE_ID); err != nil {
		err := repo.Ctx().System.SaveInt(domain.CN_UPDATE_ID, 0)
		if err != nil {
			log.LOG_Debug("SyncController::Start()-> SaveInt()",
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
		log.LOG_Info("SyncController::updateSyncStatus() syncStatus not changed")
		return
	}
	switch newStatus {
	case domain.OutOfSync:
		log.LOG_Info("SyncController::updateSyncStatus() OutOfSync")
	case domain.Syncing:
		log.LOG_Info("SyncController::updateSyncStatus() Syncing")
	case domain.Synced:
		log.LOG_Info("SyncController::updateSyncStatus() Synced")
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
				log.LOG_Info("SyncController::watchDog() -> sync() called")
				ctrl.sync()
			}
		case <-ctrl.stopChannel:
			log.LOG_Info("SyncController::watchDog() Stopped")
			return
		}
	}
}

func (ctrl *SyncController) sync() {
	log.LOG_Debug("SyncController::sync()",
		zap.Int64("UpdateID", ctrl.updateID),
	)
	ctrl.isSyncingLock.Lock()
	if ctrl.isSyncing {
		ctrl.isSyncingLock.Unlock()
		log.LOG_Debug("SyncController::sync() Exited already syncing")
		return
	}
	ctrl.isSyncing = true
	ctrl.isSyncingLock.Unlock()

	if ctrl.UserID == 0 {
		ctrl.isSyncing = false
		return
	}

	// make sure network is connected b4 start getUpdateDifference or snapshotSync
	for ctrl.networkCtrl.Quality() == domain.DISCONNECTED || ctrl.networkCtrl.Quality() == domain.CONNECTING {
		time.Sleep(100 * time.Millisecond)
	}

	var serverUpdateID int64
	var err error
	// try each 100ms until we get serverUpdateID from server
	//for {
	serverUpdateID, err = ctrl.getUpdateState()
	if err != nil {
		log.LOG_Debug("SyncController::sync()-> getUpdateState()",
			zap.String("Error", err.Error()),
		)
		// time.Sleep(100 * time.Millisecond)
		return
	} else {
		log.LOG_Debug("SyncController::sync()-> getUpdateState()",
			zap.Int64("serverUpdateID", serverUpdateID),
			zap.Int64("UpdateID", ctrl.updateID),
		)
		// break
	}
	//}

	if ctrl.updateID == 0 || (serverUpdateID-ctrl.updateID) > domain.SnapshotSync_Threshold {
		log.LOG_Debug("SyncController::sync()-> Snapshot sync")
		// remove all messages
		err := repo.Ctx().DropAndCreateTable(&dto.Messages{})
		if err != nil {
			log.LOG_Debug("SyncController::sync()-> DropAndCreateTable()",
				zap.String("Error", err.Error()),
			)
		}
		// Get Contacts from the server
		ctrl.getContacts()
		ctrl.updateID = serverUpdateID
		ctrl.getAllDialogs(0, 100)
		err = repo.Ctx().System.SaveInt(domain.CN_UPDATE_ID, int32(ctrl.updateID))
		if err != nil {
			log.LOG_Debug("SyncController::sync()-> SaveInt()",
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
	ctrl.isSyncing = false
	log.LOG_Debug("SyncController::sync() status : " + domain.SyncStatusName[ctrl.syncStatus])
	ctrl.updateSyncStatus(domain.Synced)
}

// SetUserID
func (ctrl *SyncController) SetUserID(userID int64) {
	ctrl.UserID = userID
	log.LOG_Debug("SyncController::SetUserID()",
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

	// this waitgroup is required cuz our callbacks will be called in UIExecuter go routine
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(1)
	//ctrl.queue.ExecuteCommand(
	ctrl.queue.ExecuteRealtimeCommand(
		uint64(domain.SequentialUniqueID()),
		msg.C_UpdateGetState,
		reqBytes,
		func() {
			defer waitGroup.Done()
			err = domain.ErrRequestTimeout
			log.LOG_Debug("SyncController.getUpdateState() Error : " + err.Error())
		},
		func(m *msg.MessageEnvelope) {
			defer waitGroup.Done()
			log.LOG_Debug("SyncController.getUpdateState() Success")
			switch m.Constructor {
			case msg.C_UpdateState:
				x := new(msg.UpdateState)
				x.Unmarshal(m.Message)
				updateID = x.UpdateID
			case msg.C_Error:
				err = domain.ServerError(m.Message)
				log.LOG_Debug(err.Error())
			}
		},
		true,
		false,
	)
	waitGroup.Wait()
	return
}

// getUpdateDifference
func (ctrl *SyncController) getUpdateDifference(minUpdateID int64) {

	log.LOG_Debug("SyncController::getUpdateDifference()")
	ctrl.isUpdatingDifferenceLock.Lock()
	if ctrl.isUpdatingDifference {
		ctrl.isUpdatingDifferenceLock.Unlock()
		log.LOG_Debug("SyncController::getUpdateDifference() Exited already updatingDifference")
		return
	}
	ctrl.isUpdatingDifference = true
	ctrl.isUpdatingDifferenceLock.Unlock()

	// if updateID is zero then wait for snapshot sync
	// and when sending requests w8 till its finish
	if ctrl.updateID == 0 {
		ctrl.isUpdatingDifference = false
		log.LOG_Debug("SyncController::getUpdateDifference() Exited UpdateID is zero need snapshot sync")
		return
	}

	for minUpdateID > ctrl.updateID {
		limit := minUpdateID - ctrl.updateID
		limit++
		if limit > 100 {
			limit = 100
		}
		log.LOG_Debug("SyncController::getUpdateDifference() Entered loop",
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
			uint64(domain.SequentialUniqueID()),
			msg.C_UpdateGetDifference,
			reqBytes,
			func() {
				log.LOG_Debug("SyncController::getUpdateDifference() -> ExecuteRealtimeCommand() Timeout")
			},
			func(m *msg.MessageEnvelope) {
				ctrl.onGetDiffrenceSucceed(m)
				log.LOG_Debug("SyncController::getUpdateDifference() -> ExecuteRealtimeCommand() Success")
			},
			true,
			false,
		)
		log.LOG_Debug("SyncController::getUpdateDifference() Loop next")

	}
	log.LOG_Debug("SyncController::getUpdateDifference() Loop Finished")
	ctrl.isUpdatingDifference = false
	ctrl.updateSyncStatus(domain.Synced)
}

func (ctrl *SyncController) onGetDiffrenceSucceed(m *msg.MessageEnvelope) {
	switch m.Constructor {
	case msg.C_UpdateDifference:
		x := new(msg.UpdateDifference)
		err := x.Unmarshal(m.Message)
		if err != nil {
			log.LOG_Debug("SyncController::onGetDiffrenceSucceed()-> Unmarshal()",
				zap.String("Error", err.Error()),
			)
			return
		}
		updContainer := new(msg.UpdateContainer)
		updContainer.Updates = make([]*msg.UpdateEnvelope, 0)
		updContainer.Users = x.Users
		updContainer.Groups = x.Groups
		updContainer.MaxUpdateID = x.MaxUpdateID
		updContainer.MinUpdateID = x.MinUpdateID

		log.LOG_Warn("SyncController::onGetDiffrenceSucceed()",
			zap.Int64("UpdateID", ctrl.updateID),
			zap.Int64("MaxUpdateID", x.MaxUpdateID),
			zap.Int64("MinUpdateID", x.MinUpdateID),
		)

		// No need to wait here till DB gets synced cuz UI will have required data
		go func() {
			// Save Groups
			repo.Ctx().Groups.SaveMany(x.Groups)
			// Save Users
			repo.Ctx().Users.SaveMany(x.Users)
		}()

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

			log.LOG_Debug("SyncController::onGetDiffrenceSucceed() loop",
				zap.String("Constructor", msg.ConstructorNames[update.Constructor]),
			)

			if applier, ok := ctrl.updateAppliers[update.Constructor]; ok {

				// TODO: hotfix added this cuz sync works awkwardly
				externalHandlerUpdates := applier(update)
				// check updateID to ensure to do not pass old updates to UI.
				// TODO : server still sends typing updates
				if update.UpdateID > ctrl.updateID && update.Constructor != msg.C_UpdateUserTyping {
					updContainer.Updates = append(updContainer.Updates, externalHandlerUpdates...)
				}
			}
		}
		updContainer.Length = int32(len(updContainer.Updates))
		// wrapped to UpdateContainer
		buff, _ := updContainer.Marshal()
		cmd.GetUIExecuter().Exec(func() {
			if ctrl.onUpdateMainDelegate != nil {
				ctrl.onUpdateMainDelegate(msg.C_UpdateContainer, buff)
			}
		})

		// update last updateID
		if ctrl.updateID < x.MaxUpdateID {
			ctrl.updateID = x.MaxUpdateID

			// Save UpdateID to DB
			err := repo.Ctx().System.SaveInt(domain.CN_UPDATE_ID, int32(ctrl.updateID))
			if err != nil {
				log.LOG_Debug("SyncController::onGetDiffrenceSucceed()-> SaveInt()",
					zap.String("Error", err.Error()),
				)
			}
		}
	case msg.C_Error:
		log.LOG_Debug("SyncController::onGetDiffrenceSucceed()-> C_Error",
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
		uint64(domain.SequentialUniqueID()),
		msg.C_ContactsGet,
		reqBytes,
		nil,
		func(m *msg.MessageEnvelope) {
			// SyncController applier will take care of this
		},
		false,
	)
}

// getAllDialogs
func (ctrl *SyncController) getAllDialogs(offset int32, limit int32) {
	log.LOG_Info("SyncController::getAllDialogs()")
	req := new(msg.MessagesGetDialogs)
	req.Limit = limit
	req.Offset = offset
	reqBytes, _ := req.Marshal()
	ctrl.queue.ExecuteCommand(
		uint64(domain.SequentialUniqueID()),
		msg.C_MessagesGetDialogs,
		reqBytes,
		func() {
			log.LOG_Info("SyncController::getAllDialogs() -> onTimeoutback() retry to getAllDialogs()")
			ctrl.getAllDialogs(offset, limit)
		},
		func(m *msg.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_MessagesDialogs:
				x := new(msg.MessagesDialogs)
				err := x.Unmarshal(m.Message)
				if err != nil {
					log.LOG_Info("SyncController::getAllDialogs() -> onSuccessCallback() -> Unmarshal() ",
						zap.String("Error", err.Error()),
					)
					return
				}
				log.LOG_Debug("SyncController::getAllDialogs() -> onSuccessCallback() -> MessagesDialogs",
					zap.Int("DialogsLength", len(x.Dialogs)),
					zap.Int32("Offset", offset),
					zap.Int32("Total", x.Count),
				)
				mMessages := make(map[int64]*msg.UserMessage)
				for _, message := range x.Messages {
					err := repo.Ctx().Messages.SaveMessage(message)
					if err != nil {
						log.LOG_Info("SyncController::getAllDialogs() -> onSuccessCallback() -> SaveMessage() ",
							zap.String("Error", err.Error()),
						)
					}

					mMessages[message.ID] = message
				}

				for _, dialog := range x.Dialogs {
					topMessage, _ := mMessages[dialog.TopMessageID]
					if topMessage == nil {
						log.LOG_Info("SyncController::getAllDialogs() -> onSuccessCallback() -> dialog TopMessage is null ",
							zap.Int64("MessageID", dialog.TopMessageID),
						)
						continue
					}
					err := repo.Ctx().Dialogs.SaveDialog(dialog, topMessage.CreatedOn)
					if err != nil {
						log.LOG_Info("SyncController::getAllDialogs() -> onSuccessCallback() -> SaveDialog() ",
							zap.String("Error", err.Error()),
							zap.String("Dialog", fmt.Sprintf("%v", dialog)),
						)
					}

					// create MessageHole
					err = createMessageHole(dialog.PeerID, 0, dialog.TopMessageID)
					if err != nil {
						log.LOG_Info("SyncController::getAllDialogs() -> createMessageHole() ",
							zap.String("Error", err.Error()),
						)
					}
				}
				for _, user := range x.Users {
					err := repo.Ctx().Users.SaveUser(user)
					if err != nil {
						log.LOG_Info("SyncController::getAllDialogs() -> onSuccessCallback() -> SaveUser() ",
							zap.String("Error", err.Error()),
							zap.String("User", fmt.Sprintf("%v", user)),
						)
					}
				}
				for _, group := range x.Groups {
					err := repo.Ctx().Groups.Save(group)
					if err != nil {
						log.LOG_Info("SyncController::getAllDialogs() -> onSuccessCallback() -> Groups.Save() ",
							zap.String("Error", err.Error()),
							zap.String("Group", fmt.Sprintf("%v", group)),
						)
					}
				}
				if x.Count > offset+limit {
					log.LOG_Info("SyncController::getAllDialogs() -> onSuccessCallback() retry to getAllDialogs()",
						zap.Int32("x.Count", x.Count),
						zap.Int32("offset+limit", offset+limit),
					)
					ctrl.getAllDialogs(offset+limit, limit)
				}
			case msg.C_Error:
				log.LOG_Debug("SyncController::onSuccessCallback()-> C_Error",
					zap.String("Error", domain.ServerError(m.Message).Error()),
				)
			}
		},
		false,
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
		log.LOG_Debug("SyncController::MessageHandler() Received",
			zap.String("Constructor", msg.ConstructorNames[m.Constructor]),
		)
		if applier, ok := ctrl.messageAppliers[m.Constructor]; ok {
			applier(m)
			log.LOG_Debug("SyncController::MessageHandler() Message Applied",
				zap.String("Constructor", msg.ConstructorNames[m.Constructor]),
			)
		}
	}

}

// UpdateHandler
func (ctrl *SyncController) UpdateHandler(u *msg.UpdateContainer) {

	if ctrl.syncStatus != domain.Synced {
		log.LOG_Debug("SyncController::UpdateHandler() Ignore updates while syncing")
		return
	}
	log.LOG_Debug("SyncController::UpdateHandler() Called",
		zap.Int64("UpdateID", ctrl.updateID),
		zap.Int64("MaxID", u.MaxUpdateID),
		zap.Int64("MinID", u.MinUpdateID),
		zap.Int("Count : ", len(u.Updates)),
	)
	ctrl.lastUpdateReceived = time.Now()
	if u.MinUpdateID != 0 {
		// Check if we are out of sync with server, if yes, then get the difference and
		// try to sync with server again
		if ctrl.updateID < u.MinUpdateID-1 && !ctrl.isUpdatingDifference {
			log.LOG_Debug("SyncController::UpdateHandler() calling getUpdateDifference()",
				zap.Int64("UpdateID", ctrl.updateID),
				zap.Int64("MinUpdateID", u.MinUpdateID),
			)
			ctrl.updateSyncStatus(domain.OutOfSync)
			ctrl.getUpdateDifference(u.MinUpdateID)
		}

		if ctrl.updateID < u.MaxUpdateID {
			ctrl.updateID = u.MaxUpdateID
			err := repo.Ctx().System.SaveInt(domain.CN_UPDATE_ID, int32(ctrl.updateID))
			if err != nil {
				log.LOG_Debug("SyncController::UpdateHandler() -> SaveInt()",
					zap.String("Error", err.Error()),
				)
			}
		}
	}

	udpContainer := new(msg.UpdateContainer)
	udpContainer.Updates = make([]*msg.UpdateEnvelope, 0)
	udpContainer.MaxUpdateID = u.MaxUpdateID
	udpContainer.MinUpdateID = u.MinUpdateID
	udpContainer.Users = u.Users
	udpContainer.Groups = u.Groups

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

		log.LOG_Debug("SyncController::UpdateHandler() Update Received",
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

		var externalHandlerUpdates []*msg.UpdateEnvelope
		if applier, ok := ctrl.updateAppliers[update.Constructor]; ok {

			externalHandlerUpdates = applier(update)
			log.LOG_Debug("SyncController::UpdateHandler() Update Applied",
				zap.Int64("UPDATE_ID", update.UpdateID),
				zap.String("Constructor", msg.ConstructorNames[update.Constructor]),
			)
		} else {
			// add update if not in update appliers
			udpContainer.Updates = append(udpContainer.Updates, update)
		}

		if externalHandlerUpdates != nil {
			udpContainer.Updates = append(udpContainer.Updates, externalHandlerUpdates...)

		} else {
			log.LOG_Debug("SyncController::UpdateHandler() Do not pass update to external handler",
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
			if ctrl.onUpdateMainDelegate != nil {
				ctrl.onUpdateMainDelegate(msg.C_UpdateContainer, buff)
			}
		})
	}

}

func (ctrl *SyncController) UpdateID() int64 {
	return ctrl.updateID
}

func (ctrl *SyncController) CheckSyncState() {
	go ctrl.sync()
}

func (ctrl *SyncController) ClearUpdateID() {
	ctrl.updateID = 0
}
