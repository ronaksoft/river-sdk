package synchronizer

import (
	"fmt"
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/pkg/filemanager"
	"git.ronaksoftware.com/ronak/riversdk/pkg/uiexec"

	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/network"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"

	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/queue"
	"go.uber.org/zap"
)

// Config sync controller required configs
type Config struct {
	ConnInfo    domain.RiverConfiger
	NetworkCtrl *network.Controller
	QueueCtrl   *queue.Controller
}

// Controller cache received data from server to client DB
type Controller struct {
	connInfo domain.RiverConfiger
	//sync.Mutex
	networkCtrl          *network.Controller
	queue                *queue.Controller
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

// NewSyncController create new instance
func NewSyncController(config Config) *Controller {
	ctrl := new(Controller)
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

		msg.C_UpdateGroupParticipantAdmin: ctrl.updateGroupParticipantAdmin,
		msg.C_UpdateReadMessagesContents:  ctrl.updateReadMessagesContents,

		msg.C_UpdateUserPhoto:  ctrl.updateUserPhoto,
		msg.C_UpdateGroupPhoto: ctrl.updateGroupPhoto,
		msg.C_UpdateTooLong:    ctrl.updateTooLong,
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

// Start controller
func (ctrl *Controller) Start() {
	logs.Info("Start")

	// Load the latest UpdateID stored in DB
	if v, err := repo.Ctx().System.LoadInt(domain.ColumnUpdateID); err != nil {
		err := repo.Ctx().System.SaveInt(domain.ColumnUpdateID, 0)
		if err != nil {
			logs.Error("Start()-> SaveInt()", zap.Error(err))
		}
		ctrl.updateID = 0
	} else {
		ctrl.updateID = int64(v)
	}

	// Sync with Server
	go ctrl.sync()

	go ctrl.watchDog()
}

// Stop controller
func (ctrl *Controller) Stop() {
	ctrl.stopChannel <- true // for watchDog()
}

// SetSyncStatusChangedCallback status change delegate/callback
func (ctrl *Controller) SetSyncStatusChangedCallback(h domain.SyncStatusUpdateCallback) {
	ctrl.onSyncStatusChange = h
}

// SetOnUpdateCallback set delegate to pass updates that received by getDifference to UI
func (ctrl *Controller) SetOnUpdateCallback(h domain.OnUpdateMainDelegateHandler) {
	ctrl.onUpdateMainDelegate = h
}

// updateSyncStatus
func (ctrl *Controller) updateSyncStatus(newStatus domain.SyncStatus) {
	if ctrl.syncStatus == newStatus {
		logs.Debug("updateSyncStatus() syncStatus not changed", zap.String("status", domain.SyncStatusName[newStatus]))
		return
	}
	switch newStatus {
	case domain.OutOfSync:
		logs.Info("updateSyncStatus() OutOfSync")
	case domain.Syncing:
		logs.Info("updateSyncStatus() Syncing")
	case domain.Synced:
		logs.Info("updateSyncStatus() Synced")
	}
	ctrl.syncStatus = newStatus

	if ctrl.onSyncStatusChange != nil {
		ctrl.onSyncStatusChange(newStatus)
	}
}

// watchDog
// Checks if we have not received any updates since last watch tries to re-sync with server.
func (ctrl *Controller) watchDog() {
	for {
		select {
		case <-time.After(30 * time.Second):
			// make sure network is connected b4 start getUpdateDifference or snapshotSync
			for ctrl.networkCtrl.Quality() == domain.NetworkDisconnected || ctrl.networkCtrl.Quality() == domain.NetworkConnecting {
				time.Sleep(100 * time.Millisecond)
			}
			if ctrl.syncStatus != domain.Syncing {
				logs.Info("watchDog() -> sync() called")
				ctrl.sync()
			}
		case <-ctrl.stopChannel:
			logs.Warn("watchDog() Stopped")
			return
		}
	}
}

func (ctrl *Controller) sync() {
	logs.Debug("sync()",
		zap.Int64("UpdateID", ctrl.updateID),
	)
	ctrl.isSyncingLock.Lock()
	if ctrl.isSyncing {
		ctrl.isSyncingLock.Unlock()
		logs.Debug("sync() Exited already syncing")
		return
	}
	ctrl.isSyncing = true
	ctrl.isSyncingLock.Unlock()

	if ctrl.UserID == 0 {
		ctrl.isSyncing = false
		return
	}

	// make sure network is connected b4 start getUpdateDifference or snapshotSync
	for ctrl.networkCtrl.Quality() == domain.NetworkDisconnected || ctrl.networkCtrl.Quality() == domain.NetworkConnecting {
		time.Sleep(100 * time.Millisecond)
	}

	var serverUpdateID int64
	var err error

	// get updateID from server
	serverUpdateID, err = ctrl.getUpdateState()
	if err != nil {
		logs.Error("sync()-> getUpdateState()", zap.Error(err))
		ctrl.isSyncing = false
		return
	}

	logs.Debug("sync()-> getUpdateState()",
		zap.Int64("serverUpdateID", serverUpdateID),
		zap.Int64("UpdateID", ctrl.updateID),
	)

	if ctrl.updateID == 0 || (serverUpdateID-ctrl.updateID) > domain.SnapshotSyncThreshold {
		logs.Debug("sync()-> Snapshot sync")
		// remove all messages
		err := repo.Ctx().DropAndCreateTable(&dto.Messages{})
		if err != nil {
			logs.Error("sync()-> DropAndCreateTable()", zap.Error(err))
		}
		// Get Contacts from the server
		ctrl.getContacts()
		ctrl.updateID = serverUpdateID
		ctrl.getAllDialogs(0, 100)
		err = repo.Ctx().System.SaveInt(domain.ColumnUpdateID, int32(ctrl.updateID))
		if err != nil {
			logs.Error("sync()-> SaveInt()", zap.Error(err))
		}
	} else if time.Now().Sub(ctrl.lastUpdateReceived).Truncate(time.Second) > 30 {
		// if it is passed over 60 seconds from the last update received it fetches the update
		// difference from the server
		if serverUpdateID > ctrl.updateID+1 {
			ctrl.updateSyncStatus(domain.OutOfSync)
			ctrl.getUpdateDifference(serverUpdateID + 1) // +1 cuz in here we dont have serverUpdateID itself too
		}
	}
	ctrl.isSyncing = false
	logs.Debug("sync() status : " + domain.SyncStatusName[ctrl.syncStatus])
	ctrl.updateSyncStatus(domain.Synced)
}

// SetUserID set Controller userID
func (ctrl *Controller) SetUserID(userID int64) {
	ctrl.UserID = userID
	logs.Debug("SetUserID()",
		zap.Int64("UserID", userID),
	)
}

// getUpdateState responsibility is to only get server updateID
func (ctrl *Controller) getUpdateState() (updateID int64, err error) {
	updateID = 0
	// when network is disconnected no need to enqueue update request in goque
	if ctrl.networkCtrl.Quality() == domain.NetworkDisconnected || ctrl.networkCtrl.Quality() == domain.NetworkConnecting {
		return -1, domain.ErrNoConnection
	}

	req := new(msg.UpdateGetState)
	reqBytes, _ := req.Marshal()

	// this waitgroup is required cuz our callbacks will be called in UIExecutor go routine
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
			logs.Debug("Controller.getUpdateState() Error : " + err.Error())
		},
		func(m *msg.MessageEnvelope) {
			defer waitGroup.Done()
			logs.Debug("Controller.getUpdateState() Success")
			switch m.Constructor {
			case msg.C_UpdateState:
				x := new(msg.UpdateState)
				x.Unmarshal(m.Message)
				updateID = x.UpdateID
			case msg.C_Error:
				err = domain.ParseServerError(m.Message)
				logs.Debug(err.Error())
			}
		},
		true,
		false,
	)
	waitGroup.Wait()
	return
}

// getUpdateDifference
func (ctrl *Controller) getUpdateDifference(minUpdateID int64) {

	logs.Debug("getUpdateDifference()")

	ctrl.isUpdatingDifferenceLock.Lock()
	if ctrl.isUpdatingDifference {
		ctrl.isUpdatingDifferenceLock.Unlock()
		logs.Debug("getUpdateDifference() Exited already updatingDifference")
		return
	}
	ctrl.isUpdatingDifference = true
	ctrl.isUpdatingDifferenceLock.Unlock()

	// if updateID is zero then wait for snapshot sync
	// and when sending requests w8 till its finish
	if ctrl.updateID == 0 && minUpdateID > domain.SnapshotSyncThreshold {
		ctrl.isUpdatingDifference = false
		logs.Debug("getUpdateDifference() Exited UpdateID is zero need snapshot sync")
		return
	}

	loopRepeatCounter := 0
	for minUpdateID > ctrl.updateID {
		loopRepeatCounter++

		fromUpdateID := ctrl.updateID + 1 // cuz we already have updateID itself
		limit := minUpdateID - fromUpdateID
		if limit > 100 {
			limit = 100
		}
		if limit <= 0 {
			break
		}

		logs.Debug("getUpdateDifference() Entered loop",
			zap.Int("LoopRepeatCounter", loopRepeatCounter),
			zap.Int64("limit", limit),
			zap.Int64("updateID", ctrl.updateID),
			zap.Int64("minUpdateID", minUpdateID),
		)

		ctrl.updateSyncStatus(domain.Syncing)
		req := new(msg.UpdateGetDifference)
		req.Limit = int32(limit)
		req.From = fromUpdateID
		reqBytes, _ := req.Marshal()
		ctrl.queue.ExecuteRealtimeCommand(
			uint64(domain.SequentialUniqueID()),
			msg.C_UpdateGetDifference,
			reqBytes,
			func() {
				logs.Debug("getUpdateDifference() -> ExecuteRealtimeCommand() Timeout")
			},
			func(m *msg.MessageEnvelope) {
				ctrl.onGetDiffrenceSucceed(m)
				logs.Debug("getUpdateDifference() -> ExecuteRealtimeCommand() Success")
			},
			true,
			false,
		)
		logs.Debug("getUpdateDifference() Loop next")

	}
	logs.Debug("getUpdateDifference() Loop Finished")
	ctrl.isUpdatingDifference = false
	ctrl.updateSyncStatus(domain.Synced)
}

func (ctrl *Controller) onGetDiffrenceSucceed(m *msg.MessageEnvelope) {
	switch m.Constructor {
	case msg.C_UpdateDifference:
		x := new(msg.UpdateDifference)
		err := x.Unmarshal(m.Message)
		if err != nil {
			logs.Error("onGetDiffrenceSucceed()-> Unmarshal()", zap.Error(err))
			return
		}
		updContainer := new(msg.UpdateContainer)
		updContainer.Updates = make([]*msg.UpdateEnvelope, 0)
		updContainer.Users = x.Users
		updContainer.Groups = x.Groups
		updContainer.MaxUpdateID = x.MaxUpdateID
		updContainer.MinUpdateID = x.MinUpdateID

		logs.Info("onGetDiffrenceSucceed()",
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

			logs.Debug("onGetDiffrenceSucceed() loop",
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

		// update last updateID
		if ctrl.updateID < x.MaxUpdateID {
			ctrl.updateID = x.MaxUpdateID

			// Save UpdateID to DB
			err := repo.Ctx().System.SaveInt(domain.ColumnUpdateID, int32(ctrl.updateID))
			if err != nil {
				logs.Error("onGetDiffrenceSucceed()-> SaveInt()", zap.Error(err))
			}
		}

		// wrapped to UpdateContainer
		buff, _ := updContainer.Marshal()
		uiexec.Ctx().Exec(func() {
			if ctrl.onUpdateMainDelegate != nil {
				ctrl.onUpdateMainDelegate(msg.C_UpdateContainer, buff)
			}
		})

	case msg.C_Error:
		logs.Debug("onGetDiffrenceSucceed()-> C_Error",
			zap.String("Error", domain.ParseServerError(m.Message).Error()),
		)
		// TODO:: Handle error
	}
}

// getContacts
func (ctrl *Controller) getContacts() {
	req := new(msg.ContactsGet)
	reqBytes, _ := req.Marshal()
	ctrl.queue.ExecuteCommand(
		uint64(domain.SequentialUniqueID()),
		msg.C_ContactsGet,
		reqBytes,
		nil,
		func(m *msg.MessageEnvelope) {
			// Controller applier will take care of this
		},
		false,
	)
}

// getAllDialogs
func (ctrl *Controller) getAllDialogs(offset int32, limit int32) {
	logs.Info("getAllDialogs()")
	req := new(msg.MessagesGetDialogs)
	req.Limit = limit
	req.Offset = offset
	reqBytes, _ := req.Marshal()
	ctrl.queue.ExecuteCommand(
		uint64(domain.SequentialUniqueID()),
		msg.C_MessagesGetDialogs,
		reqBytes,
		func() {
			logs.Warn("getAllDialogs() -> onTimeoutback() retry to getAllDialogs()")
			ctrl.getAllDialogs(offset, limit)
		},
		func(m *msg.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_MessagesDialogs:
				x := new(msg.MessagesDialogs)
				err := x.Unmarshal(m.Message)
				if err != nil {
					logs.Error("getAllDialogs() -> onSuccessCallback() -> Unmarshal() ", zap.Error(err))
					return
				}
				logs.Debug("getAllDialogs() -> onSuccessCallback() -> MessagesDialogs",
					zap.Int("DialogsLength", len(x.Dialogs)),
					zap.Int32("Offset", offset),
					zap.Int32("Total", x.Count),
				)
				mMessages := make(map[int64]*msg.UserMessage)
				for _, message := range x.Messages {
					err := repo.Ctx().Messages.SaveMessage(message)
					if err != nil {
						logs.Error("getAllDialogs() -> onSuccessCallback() -> SaveMessage() ", zap.Error(err))
					}

					mMessages[message.ID] = message
				}

				for _, dialog := range x.Dialogs {
					topMessage, _ := mMessages[dialog.TopMessageID]
					if topMessage == nil {
						logs.Error("getAllDialogs() -> onSuccessCallback() -> dialog TopMessage is null ",
							zap.Int64("MessageID", dialog.TopMessageID),
						)
						continue
					}
					// create MessageHole
					err = CreateMessageHole(dialog.PeerID, 0, dialog.TopMessageID-1)
					if err != nil {
						logs.Error("getAllDialogs() -> createMessageHole() ", zap.Error(err))
					}
					// make sure to created the messagehole b4 creating dialog
					err := repo.Ctx().Dialogs.SaveDialog(dialog, topMessage.CreatedOn)
					if err != nil {
						logs.Error("getAllDialogs() -> onSuccessCallback() -> SaveDialog() ",
							zap.String("Error", err.Error()),
							zap.String("Dialog", fmt.Sprintf("%v", dialog)),
						)
					}
				}

				for _, user := range x.Users {
					err := repo.Ctx().Users.SaveUser(user)
					if err != nil {
						logs.Error("getAllDialogs() -> onSuccessCallback() -> SaveUser() ",
							zap.String("Error", err.Error()),
							zap.String("User", fmt.Sprintf("%v", user)),
						)
					}
				}
				for _, group := range x.Groups {
					err := repo.Ctx().Groups.Save(group)
					if err != nil {
						logs.Error("getAllDialogs() -> onSuccessCallback() -> Groups.Save() ",
							zap.String("Error", err.Error()),
							zap.String("Group", fmt.Sprintf("%v", group)),
						)
					}
				}
				if x.Count > offset+limit {
					logs.Info("getAllDialogs() -> onSuccessCallback() retry to getAllDialogs()",
						zap.Int32("x.Count", x.Count),
						zap.Int32("offset+limit", offset+limit),
					)
					ctrl.getAllDialogs(offset+limit, limit)
				}
			case msg.C_Error:
				logs.Error("onSuccessCallback()-> C_Error",
					zap.String("Error", domain.ParseServerError(m.Message).Error()),
				)
			}
		},
		false,
	)

}

func (ctrl *Controller) addToDeliveredMessageList(id int64) {
	ctrl.deliveredMessagesMutex.Lock()
	ctrl.deliveredMessages[id] = true
	ctrl.deliveredMessagesMutex.Unlock()
}

func (ctrl *Controller) isDeliveredMessage(id int64) bool {
	ctrl.deliveredMessagesMutex.Lock()
	var ok bool
	if _, ok = ctrl.deliveredMessages[id]; ok {
		// cuz server sends duplicated updates again do not remove deliveredMessages
		// delete(ctrl.deliveredMessages, id)
	}
	ctrl.deliveredMessagesMutex.Unlock()

	return ok
}

// Status displays SyncStatus
func (ctrl *Controller) Status() domain.SyncStatus {
	return ctrl.syncStatus
}

// MessageHandler call appliers-> repository and sync data
func (ctrl *Controller) MessageHandler(messages []*msg.MessageEnvelope) {
	for _, m := range messages {
		logs.Debug("MessageHandler() Received",
			zap.String("Constructor", msg.ConstructorNames[m.Constructor]),
		)

		if m.Constructor == msg.C_Error {
			err := new(msg.Error)
			err.Unmarshal(m.Message)
			logs.Error("MessageHandler() Received Error ", zap.String("Code", err.Code), zap.String("Item", err.Items))
		}

		if applier, ok := ctrl.messageAppliers[m.Constructor]; ok {
			applier(m)
			logs.Debug("MessageHandler() Message Applied",
				zap.String("Constructor", msg.ConstructorNames[m.Constructor]),
			)
		}
	}

}

// UpdateHandler receives update to cache them in client DB
func (ctrl *Controller) UpdateHandler(u *msg.UpdateContainer) {

	logs.Debug("UpdateHandler() Called",
		zap.Int64("ctrl.UpdateID", ctrl.updateID),
		zap.Int64("MaxID", u.MaxUpdateID),
		zap.Int64("MinID", u.MinUpdateID),
		zap.Int("Count : ", len(u.Updates)),
	)
	ctrl.lastUpdateReceived = time.Now()
	if u.MinUpdateID != 0 {
		// Check if we are out of sync with server, if yes, then get the difference and
		// try to sync with server again
		if ctrl.updateID < u.MinUpdateID-1 && !ctrl.isUpdatingDifference {
			logs.Debug("UpdateHandler() calling getUpdateDifference()",
				zap.Int64("UpdateID", ctrl.updateID),
				zap.Int64("MinUpdateID", u.MinUpdateID),
			)
			ctrl.updateSyncStatus(domain.OutOfSync)
			ctrl.getUpdateDifference(u.MinUpdateID)
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

		logs.Debug("UpdateHandler() Update Received",
			zap.String("Constructor", msg.ConstructorNames[update.Constructor]),
		)

		var externalHandlerUpdates []*msg.UpdateEnvelope
		if applier, ok := ctrl.updateAppliers[update.Constructor]; ok {

			externalHandlerUpdates = applier(update)
			logs.Info("UpdateHandler() Update Applied",
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
			logs.Warn("UpdateHandler() Do not pass update to external handler",
				zap.Int64("UPDATE_ID", update.UpdateID),
				zap.String("Constructor", msg.ConstructorNames[update.Constructor]),
			)
		}

	}

	// save updateID after processing messages
	if ctrl.updateID < u.MaxUpdateID {
		ctrl.updateID = u.MaxUpdateID
		err := repo.Ctx().System.SaveInt(domain.ColumnUpdateID, int32(ctrl.updateID))
		if err != nil {
			logs.Error("UpdateHandler() -> SaveInt()", zap.Error(err))
		}
	}

	udpContainer.Length = int32(len(udpContainer.Updates))

	// call external handler
	if ctrl.onUpdateMainDelegate != nil {

		// wrapped to UpdateContainer
		buff, _ := udpContainer.Marshal()

		// pass all updates to UI
		uiexec.Ctx().Exec(func() {
			if ctrl.onUpdateMainDelegate != nil {
				ctrl.onUpdateMainDelegate(msg.C_UpdateContainer, buff)
			}
		})
	}

}

// UpdateID returns current updateID
func (ctrl *Controller) UpdateID() int64 {
	return ctrl.updateID
}

// CheckSyncState enforce to check client updateID with server getState updateID
func (ctrl *Controller) CheckSyncState() {
	go ctrl.sync()
}

// ClearUpdateID reset updateID
func (ctrl *Controller) ClearUpdateID() {
	ctrl.updateID = 0
	ctrl.UserID = 0
}

// handleMediaMessage extract files info from messages that have Document object
func handleMediaMessage(messages ...*msg.UserMessage) {
	for _, m := range messages {
		switch m.MediaType {
		case msg.MediaTypeEmpty:
			// NOP
		case msg.MediaTypePhoto:
			logs.Info("handleMediaMessage() Message.SharedMediaType is msg.MediaTypePhoto")
			// TODO:: implement it
		case msg.MediaTypeDocument:
			mediaDoc := new(msg.MediaDocument)
			err := mediaDoc.Unmarshal(m.Media)
			if err == nil {
				repo.Ctx().Files.SaveFileDocument(m, mediaDoc)
				t := mediaDoc.Doc.Thumbnail
				if t != nil {
					if t.FileID != 0 {
						go filemanager.Ctx().DownloadThumbnail(m.ID, t.FileID, t.AccessHash, t.ClusterID, 0)
					}
				}

			} else {
				logs.Error("handleMediaMessage()-> connat unmarshal MediaTypeDocument", zap.Error(err))
			}
		case msg.MediaTypeContact:
			logs.Info("handleMediaMessage() Message.SharedMediaType is msg.MediaTypeContact")
			// TODO:: implement it
		default:
			logs.Info("handleMediaMessage() Message.SharedMediaType is invalid")
		}
	}
}

// ContactImportFromServer import contact from server
func (ctrl *Controller) ContactImportFromServer() {
	contactsGetHash, err := repo.Ctx().System.LoadInt(domain.ColumnContactsGetHash)
	if err != nil {
		logs.Error("onNetworkControllerConnected() failed to get contactsGetHash", zap.Error(err))
	}
	contactGetReq := new(msg.ContactsGet)
	contactGetReq.Crc32Hash = uint32(contactsGetHash)
	contactGetBytes, _ := contactGetReq.Marshal()
	ctrl.queue.ExecuteRealtimeCommand(uint64(domain.SequentialUniqueID()), msg.C_ContactsGet, contactGetBytes, nil, nil, false, false)
}
