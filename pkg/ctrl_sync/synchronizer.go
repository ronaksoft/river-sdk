package synchronizer

import (
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/pkg/filemanager"
	messageHole "git.ronaksoftware.com/ronak/riversdk/pkg/message_hole"
	"sync"
	"sync/atomic"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/pkg/uiexec"

	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_network"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_queue"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"go.uber.org/zap"
)

// Config sync controller required configs
type Config struct {
	ConnInfo    domain.RiverConfigurator
	NetworkCtrl *network.Controller
	QueueCtrl   *queue.Controller
}

// Controller cache received data from server to client DB
type Controller struct {
	connInfo             domain.RiverConfigurator
	networkCtrl          *network.Controller
	queueCtrl            *queue.Controller
	onSyncStatusChange   domain.SyncStatusUpdateCallback
	onUpdateMainDelegate domain.OnUpdateMainDelegateHandler
	syncStatus           domain.SyncStatus
	lastUpdateReceived   time.Time
	updateID             int64
	updateAppliers       map[int64]domain.UpdateApplier
	messageAppliers      map[int64]domain.MessageApplier
	stopChannel          chan bool
	userID               int64

	// delivered Message
	deliveredMessagesMutex sync.Mutex
	deliveredMessages      map[int64]bool

	// internal locks
	updateDifferenceLock int32
	syncLock             int32
}

// NewSyncController create new instance
func NewSyncController(config Config) *Controller {
	ctrl := new(Controller)
	ctrl.stopChannel = make(chan bool)
	ctrl.connInfo = config.ConnInfo
	ctrl.queueCtrl = config.QueueCtrl
	ctrl.networkCtrl = config.NetworkCtrl

	ctrl.updateAppliers = map[int64]domain.UpdateApplier{
		msg.C_UpdateNewMessage:            ctrl.updateNewMessage,
		msg.C_UpdateReadHistoryOutbox:     ctrl.updateReadHistoryOutbox,
		msg.C_UpdateReadHistoryInbox:      ctrl.updateReadHistoryInbox,
		msg.C_UpdateMessageEdited:         ctrl.updateMessageEdited,
		msg.C_UpdateMessageID:             ctrl.updateMessageID,
		msg.C_UpdateNotifySettings:        ctrl.updateNotifySettings,
		msg.C_UpdateUsername:              ctrl.updateUsername,
		msg.C_UpdateMessagesDeleted:       ctrl.updateMessagesDeleted,
		msg.C_UpdateGroupParticipantAdmin: ctrl.updateGroupParticipantAdmin,
		msg.C_UpdateReadMessagesContents:  ctrl.updateReadMessagesContents,
		msg.C_UpdateUserPhoto:             ctrl.updateUserPhoto,
		msg.C_UpdateGroupPhoto:            ctrl.updateGroupPhoto,
		msg.C_UpdateTooLong:               ctrl.updateTooLong,
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

// watchDog
// Checks if we have not received any updates since last watch tries to re-sync with server.
func (ctrl *Controller) watchDog() {
	syncTime := 60 * time.Second
	for {
		select {
		case <-time.After(syncTime):
			// Wait for network
			ctrl.networkCtrl.WaitForNetwork()

			// Check if we were not syncing in the last 60 seconds
			if time.Now().Sub(ctrl.lastUpdateReceived) < syncTime {
				break
			}
			ctrl.sync()
		case <-ctrl.stopChannel:
			logs.Info("SyncController::watchDog() Stopped")
			return
		}
	}
}

func (ctrl *Controller) sync() {
	// Check if sync function is already running, then return otherwise lock it and continue
	if !atomic.CompareAndSwapInt32(&ctrl.syncLock, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&ctrl.syncLock, 0)

	logs.Debug("SyncController::sync()",
		zap.Int64("UpdateID", ctrl.updateID),
	)
	// There is no need to sync when no user has been authorized
	if ctrl.userID == 0 {
		return
	}

	// Wait until network is available
	ctrl.networkCtrl.WaitForNetwork()

	// get updateID from server
	serverUpdateID, err := getUpdateState(ctrl)
	if err != nil {
		logs.Error("sync()-> getUpdateState()", zap.Error(err))
		return
	}
	if ctrl.updateID == serverUpdateID {
		return
	}

	// Update the sync controller status
	updateSyncStatus(ctrl, domain.Syncing)
	defer updateSyncStatus(ctrl, domain.Synced)

	if ctrl.updateID == 0 || (serverUpdateID-ctrl.updateID) > domain.SnapshotSyncThreshold {
		logs.Info("SyncController::Snapshot sync")
		// remove all messages
		err := repo.DropAndCreateTable(&dto.Messages{})
		if err != nil {
			logs.Error("sync()-> DropAndCreateTable()", zap.Error(err))
			return
		}
		// Get Contacts from the server
		getContacts(ctrl)
		ctrl.updateID = serverUpdateID
		getAllDialogs(ctrl, 0, 100)
		err = repo.System.SaveInt(domain.ColumnUpdateID, int32(ctrl.updateID))
		if err != nil {
			logs.Error("sync()-> SaveInt()", zap.Error(err))
			return
		}
	} else if serverUpdateID > ctrl.updateID+1 {
		// if it is passed over 60 seconds from the last update received it fetches the update
		// difference from the server
		logs.Info("SyncController::Sequential sync")
		getUpdateDifference(ctrl, serverUpdateID+1) // +1 cuz in here we dont have serverUpdateID itself too
	}
}
func updateSyncStatus(ctrl *Controller, newStatus domain.SyncStatus) {
	if ctrl.syncStatus == newStatus {
		return
	}
	logs.Info("Sync Controller Status Updated",
		zap.String("Status", newStatus.ToString()),
	)
	ctrl.syncStatus = newStatus

	if ctrl.onSyncStatusChange != nil {
		ctrl.onSyncStatusChange(newStatus)
	}
}
func getUpdateState(ctrl *Controller) (updateID int64, err error) {
	updateID = 0
	if !ctrl.networkCtrl.Connected() {
		return -1, domain.ErrNoConnection
	}

	req := new(msg.UpdateGetState)
	reqBytes, _ := req.Marshal()

	// this waitGroup is required cuz our callbacks will be called in UIExecutor go routine
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(1)
	_ = ctrl.queueCtrl.ExecuteRealtimeCommand(
		uint64(domain.SequentialUniqueID()),
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
				_ = x.Unmarshal(m.Message)
				updateID = x.UpdateID
			case msg.C_Error:
				err = domain.ParseServerError(m.Message)
			}
		},
		true,
		false,
	)
	waitGroup.Wait()
	return
}
func getContacts(ctrl *Controller) {
	req := new(msg.ContactsGet)
	reqBytes, _ := req.Marshal()
	ctrl.queueCtrl.ExecuteCommand(
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
func getAllDialogs(ctrl *Controller, offset int32, limit int32) {
	logs.Info("SyncController::getAllDialogs()",
		zap.Int32("Offset", offset),
		zap.Int32("Limit", limit),
	)
	req := new(msg.MessagesGetDialogs)
	req.Limit = limit
	req.Offset = offset
	reqBytes, _ := req.Marshal()
	ctrl.queueCtrl.ExecuteCommand(
		uint64(domain.SequentialUniqueID()),
		msg.C_MessagesGetDialogs,
		reqBytes,
		func() {
			logs.Warn("getAllDialogs() -> onTimeoutback() retry to getAllDialogs()")
			getAllDialogs(ctrl, offset, limit)
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
				mMessages := make(map[int64]*msg.UserMessage)
				for _, message := range x.Messages {
					err := repo.Messages.SaveMessage(message)
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
					_ = messageHole.InsertHole(dialog.PeerID, dialog.PeerType, 0, dialog.TopMessageID-1)
					_ = messageHole.SetUpperFilled(dialog.PeerID, dialog.PeerType,  dialog.TopMessageID)

					// make sure to created the message hole b4 creating dialog
					err := repo.Dialogs.SaveDialog(dialog, topMessage.CreatedOn)
					if err != nil {
						logs.Error("getAllDialogs() -> onSuccessCallback() -> SaveDialog() ",
							zap.String("Error", err.Error()),
							zap.String("Dialog", fmt.Sprintf("%v", dialog)),
						)
					}
				}

				_ = repo.Users.SaveMany(x.Users)
				_ = repo.Groups.SaveMany(x.Groups)

				if x.Count > offset+limit {
					getAllDialogs(ctrl, offset+limit, limit)
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
func getUpdateDifference(ctrl *Controller, serverUpdateID int64) {
	// Check if getUpdateDifference function is already running, then return otherwise lock it and continue
	if !atomic.CompareAndSwapInt32(&ctrl.updateDifferenceLock, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&ctrl.updateDifferenceLock, 0)

	logs.Debug("SyncController::getUpdateDifference()",
		zap.Int64("ServerUpdateID", serverUpdateID),
		zap.Int64("ClientUpdateID", ctrl.updateID),
	)

	for serverUpdateID > ctrl.updateID {
		fromUpdateID := ctrl.updateID + 1 // cuz we already have updateID itself
		limit := serverUpdateID - fromUpdateID
		if limit > 100 {
			limit = 100
		}
		if limit <= 0 {
			break
		}

		req := new(msg.UpdateGetDifference)
		req.Limit = int32(limit)
		req.From = fromUpdateID
		reqBytes, _ := req.Marshal()
		_ = ctrl.queueCtrl.ExecuteRealtimeCommand(
			uint64(domain.SequentialUniqueID()),
			msg.C_UpdateGetDifference,
			reqBytes,
			func() {
				logs.Debug("SyncController::getUpdateDifference() -> ExecuteRealtimeCommand() Timeout")
			},
			func(m *msg.MessageEnvelope) {
				onGetDifferenceSucceed(ctrl, m)
				logs.Debug("SyncController::getUpdateDifference() -> ExecuteRealtimeCommand() Success")
			},
			true,
			false,
		)
	}
}
func onGetDifferenceSucceed(ctrl *Controller, m *msg.MessageEnvelope) {
	switch m.Constructor {
	case msg.C_UpdateDifference:
		x := new(msg.UpdateDifference)
		err := x.Unmarshal(m.Message)
		if err != nil {
			logs.Error("onGetDifferenceSucceed()-> Unmarshal()", zap.Error(err))
			return
		}
		updContainer := new(msg.UpdateContainer)
		updContainer.Updates = make([]*msg.UpdateEnvelope, 0)
		updContainer.Users = x.Users
		updContainer.Groups = x.Groups
		updContainer.MaxUpdateID = x.MaxUpdateID
		updContainer.MinUpdateID = x.MinUpdateID

		logs.Info("SyncController::onGetDifferenceSucceed()",
			zap.Int64("UpdateID", ctrl.updateID),
			zap.Int64("MaxUpdateID", x.MaxUpdateID),
			zap.Int64("MinUpdateID", x.MinUpdateID),
		)

		// No need to wait here till DB gets synced cuz UI will have required data
		go func() {
			// Save Groups
			_ = repo.Groups.SaveMany(x.Groups)
			// Save Users
			_ = repo.Users.SaveMany(x.Users)
		}()

		for _, update := range x.Updates {
			if applier, ok := ctrl.updateAppliers[update.Constructor]; ok {
				externalHandlerUpdates := applier(update)
				updContainer.Updates = append(updContainer.Updates, externalHandlerUpdates...)
			}
		}
		updContainer.Length = int32(len(updContainer.Updates))

		// update last updateID
		if ctrl.updateID < x.MaxUpdateID {
			ctrl.updateID = x.MaxUpdateID

			// Save UpdateID to DB
			err := repo.System.SaveInt(domain.ColumnUpdateID, int32(ctrl.updateID))
			if err != nil {
				logs.Error("onGetDifferenceSucceed()-> SaveInt()", zap.Error(err))
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
		logs.Debug("onGetDifferenceSucceed()-> C_Error",
			zap.String("Error", domain.ParseServerError(m.Message).Error()),
		)
		// TODO:: Handle error
	}
}

func (ctrl *Controller) SetUserID(userID int64) {
	ctrl.userID = userID
	logs.Debug("SyncController::SetUserID()",
		zap.Int64("userID", userID),
	)
}

func (ctrl *Controller) GetUserID() int64 {
	return ctrl.userID
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

// Start controller
func (ctrl *Controller) Start() {
	logs.Info("SyncController::  Start")

	// Load the latest UpdateID stored in DB
	if v, err := repo.System.LoadInt(domain.ColumnUpdateID); err != nil {
		err := repo.System.SaveInt(domain.ColumnUpdateID, 0)
		if err != nil {
			logs.Error("Start()-> SaveInt()", zap.Error(err))
		}
		ctrl.updateID = 0
	} else {
		ctrl.updateID = int64(v)
	}

	// set default value to synced status
	updateSyncStatus(ctrl, domain.Synced)

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

// MessageHandler call appliers-> repository and sync data
func (ctrl *Controller) MessageHandler(messages []*msg.MessageEnvelope) {
	for _, m := range messages {
		logs.Debug("SyncController::MessageHandler() Received",
			zap.String("Constructor", msg.ConstructorNames[m.Constructor]),
		)

		switch m.Constructor {
		case msg.C_Error:
			err := new(msg.Error)
			_ = err.Unmarshal(m.Message)
			logs.Error("MessageHandler() Received Error ", zap.String("Code", err.Code), zap.String("Item", err.Items))
		}

		if applier, ok := ctrl.messageAppliers[m.Constructor]; ok {
			applier(m)
			logs.Debug("SyncController::MessageHandler() Message Applied",
				zap.String("Constructor", msg.ConstructorNames[m.Constructor]),
			)
		}
	}

}

// UpdateHandler receives update to cache them in client DB
func (ctrl *Controller) UpdateHandler(updateContainer *msg.UpdateContainer) {
	logs.Debug("SyncController::UpdateHandler() Called",
		zap.Int64("ctrl.UpdateID", ctrl.updateID),
		zap.Int64("MaxID", updateContainer.MaxUpdateID),
		zap.Int64("MinID", updateContainer.MinUpdateID),
		zap.Int("Count : ", len(updateContainer.Updates)),
	)
	ctrl.lastUpdateReceived = time.Now()

	// Check if update has been already applied
	if updateContainer.MinUpdateID != 0 && ctrl.updateID >= updateContainer.MinUpdateID {
		return
	}

	// Check if we are out of sync with server, if yes, then call the sync() function
	// We call it in blocking mode,
	if ctrl.updateID < updateContainer.MinUpdateID-1 {
		ctrl.sync()
		return
	}

	udpContainer := new(msg.UpdateContainer)
	udpContainer.Updates = make([]*msg.UpdateEnvelope, 0)
	udpContainer.MaxUpdateID = updateContainer.MaxUpdateID
	udpContainer.MinUpdateID = updateContainer.MinUpdateID
	udpContainer.Users = updateContainer.Users
	udpContainer.Groups = updateContainer.Groups
	for _, u := range updateContainer.Users {
		// Download users avatar if its not exist
		if u.Photo != nil {
			dtoPhoto := repo.Users.GetUserPhoto(u.ID, u.Photo.PhotoID)
			if dtoPhoto != nil {
				if dtoPhoto.SmallFilePath == "" || dtoPhoto.SmallFileID != u.Photo.PhotoSmall.FileID {
					go func(userID int64, photo *msg.UserPhoto) {
						_, _ = filemanager.Ctx().DownloadAccountPhoto(userID, photo, false)
					}(u.ID, u.Photo)
				}
			} else if u.Photo.PhotoID != 0 {
				go func(userID int64, photo *msg.UserPhoto) {
					_, _ = filemanager.Ctx().DownloadAccountPhoto(userID, photo, false)
				}(u.ID, u.Photo)
			}
		}
	}
	for _, g := range updateContainer.Groups {
		// Download group avatar if its not exist
		if g.Photo != nil {
			dtoGroup, err := repo.Groups.GetGroupDTO(g.ID)
			if err == nil && dtoGroup != nil {
				if dtoGroup.SmallFilePath == "" || dtoGroup.SmallFileID != g.Photo.PhotoSmall.FileID {
					go func(groupID int64, photo *msg.GroupPhoto) {
						_, _ = filemanager.Ctx().DownloadGroupPhoto(groupID, photo, false)
					}(g.ID, g.Photo)
				}
			} else if g.Photo.PhotoSmall.FileID != 0 {
				go func(groupID int64, photo *msg.GroupPhoto) {
					_, _ = filemanager.Ctx().DownloadGroupPhoto(groupID, photo, false)
				}(g.ID, g.Photo)
			}
		}
	}
	for _, update := range updateContainer.Updates {
		logs.Debug("SyncController::UpdateHandler() Update Received",
			zap.String("Constructor", msg.ConstructorNames[update.Constructor]),
		)

		// var externalHandlerUpdates []*msg.UpdateEnvelope
		applier, ok := ctrl.updateAppliers[update.Constructor]
		if ok {
			externalHandlerUpdates := applier(update)
			switch update.Constructor {
			case msg.C_UpdateMessageID:
			default:
				udpContainer.Updates = append(udpContainer.Updates, externalHandlerUpdates...)
			}
		} else {
			udpContainer.Updates = append(udpContainer.Updates, update)
		}
	}

	// No need to wait here till DB gets synced cuz UI will have required data
	go func() {
		// Save Groups
		_ = repo.Groups.SaveMany(updateContainer.Groups)
		// Save Users
		_ = repo.Users.SaveMany(updateContainer.Users)
	}()

	// save updateID after processing messages
	if ctrl.updateID < updateContainer.MaxUpdateID {
		ctrl.updateID = updateContainer.MaxUpdateID
		err := repo.System.SaveInt(domain.ColumnUpdateID, int32(ctrl.updateID))
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

// ClearUpdateID reset updateID
func (ctrl *Controller) ClearUpdateID() {
	ctrl.updateID = 0
	ctrl.userID = 0
}

// ContactImportFromServer import contact from server
func (ctrl *Controller) ContactImportFromServer() {
	contactsGetHash, err := repo.System.LoadInt(domain.ColumnContactsGetHash)
	if err != nil {
		logs.Error("onNetworkControllerConnected() failed to get contactsGetHash", zap.Error(err))
	}
	contactGetReq := new(msg.ContactsGet)
	contactGetReq.Crc32Hash = uint32(contactsGetHash)
	contactGetBytes, _ := contactGetReq.Marshal()
	_ = ctrl.queueCtrl.ExecuteRealtimeCommand(
		uint64(domain.SequentialUniqueID()),
		msg.C_ContactsGet, contactGetBytes,
		nil, nil, false, false,
	)
}

// GetSyncStatus
func (ctrl *Controller) GetSyncStatus() domain.SyncStatus {
	return ctrl.syncStatus
}

// extractMessagesMedia extract files info from messages that have Document object
func extractMessagesMedia(messages ...*msg.UserMessage) {
	for _, m := range messages {
		switch m.MediaType {
		case msg.MediaTypeEmpty:
			// NOP
		case msg.MediaTypePhoto:
			// TODO:: implement it
		case msg.MediaTypeDocument:
			mediaDoc := new(msg.MediaDocument)
			err := mediaDoc.Unmarshal(m.Media)
			if err != nil {
				logs.Error("extractMessagesMedia()-> connot unmarshal MediaTypeDocument", zap.Error(err))
				break
			}
			_ = repo.Files.SaveFileDocument(m, mediaDoc)
			t := mediaDoc.Doc.Thumbnail
			if t != nil {
				if t.FileID != 0 {
					go filemanager.Ctx().DownloadThumbnail(m.ID, t.FileID, t.AccessHash, t.ClusterID, 0)
				}
			}

		case msg.MediaTypeContact:
		default:
		}
	}
}
