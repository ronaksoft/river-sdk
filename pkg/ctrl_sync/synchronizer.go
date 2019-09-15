package syncCtrl

import (
	"encoding/json"
	"git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_file"
	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_network"
	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_queue"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	messageHole "git.ronaksoftware.com/ronak/riversdk/pkg/message_hole"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"git.ronaksoftware.com/ronak/riversdk/pkg/salt"
	"git.ronaksoftware.com/ronak/riversdk/pkg/uiexec"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
	"time"
)

// Config sync controller required configs
type Config struct {
	ConnInfo    domain.RiverConfigurator
	NetworkCtrl *networkCtrl.Controller
	QueueCtrl   *queueCtrl.Controller
	FileCtrl    *fileCtrl.Controller
}

// Controller cache received data from server to client DB
type Controller struct {
	waitGroup            sync.WaitGroup
	connInfo             domain.RiverConfigurator
	networkCtrl          *networkCtrl.Controller
	queueCtrl            *queueCtrl.Controller
	fileCtrl             *fileCtrl.Controller
	onSyncStatusChange   domain.SyncStatusUpdateCallback
	onUpdateMainDelegate domain.OnUpdateMainDelegateHandler
	syncStatus           domain.SyncStatus
	lastUpdateReceived   time.Time
	updateID             int64
	updateAppliers       map[int64]domain.UpdateApplier
	messageAppliers      map[int64]domain.MessageApplier
	stopChannel          chan bool
	userID               int64

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
	ctrl.fileCtrl = config.FileCtrl
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
		msg.C_UpdateDialogPinned:          ctrl.updateDialogPinned,
		msg.C_UpdateAccountPrivacy:        ctrl.updateAccountPrivacy,
		msg.C_UpdateDraftMessage:          ctrl.updateDraftMessage,
		msg.C_UpdateDraftMessageCleared:   ctrl.updateDraftMessageCleared,
	}

	ctrl.messageAppliers = map[int64]domain.MessageApplier{
		msg.C_AuthAuthorization: ctrl.authAuthorization,
		msg.C_ContactsImported:  ctrl.contactsImported,
		msg.C_ContactsMany:      ctrl.contactsMany,
		msg.C_MessagesDialogs:   ctrl.messagesDialogs,
		msg.C_AuthSentCode:      ctrl.authSentCode,
		msg.C_UsersMany:         ctrl.usersMany,
		msg.C_MessagesMany:      ctrl.messagesMany,
		msg.C_GroupFull:         ctrl.groupFull,
	}

	return ctrl
}

// watchDog
// Checks if we have not received any updates since last watch tries to re-sync with server.
func (ctrl *Controller) watchDog() {
	syncTime := 3 * time.Minute
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
			logs.Info("SyncController:: watchDog Stopped")
			return
		}
	}
}
func (ctrl *Controller) Sync() {
	ctrl.sync()
}
func (ctrl *Controller) sync() {
	// Check if sync function is already running, then return otherwise lock it and continue
	if !atomic.CompareAndSwapInt32(&ctrl.syncLock, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&ctrl.syncLock, 0)

	// There is no need to sync when no user has been authorized
	if ctrl.userID == 0 {
		return
	}

	// Wait until network is available
	ctrl.networkCtrl.WaitForNetwork()

	// get updateID from server
	serverUpdateID, err := getUpdateState(ctrl)
	if err != nil {
		logs.Warn("sync()-> getUpdateState()", zap.Error(err))
		return
	}
	if ctrl.updateID == serverUpdateID {
		return
	}

	// Update the sync controller status
	updateSyncStatus(ctrl, domain.Syncing)
	defer updateSyncStatus(ctrl, domain.Synced)

	if ctrl.updateID == 0 || (serverUpdateID-ctrl.updateID) > domain.SnapshotSyncThreshold {
		logs.Info("SyncController:: Snapshot sync")

		// Clear DB
		repo.DropAll()

		// Get Contacts from the server
		waitGroup := &sync.WaitGroup{}
		waitGroup.Add(1)
		go getContacts(waitGroup, ctrl)

		waitGroup.Add(1)
		go getAllDialogs(waitGroup, ctrl, 0, 100)
		waitGroup.Wait()
		ctrl.updateID = serverUpdateID
		err = repo.System.SaveInt(domain.SkUpdateID, uint64(ctrl.updateID))
		if err != nil {
			logs.Error("sync()-> SaveInt()", zap.Error(err))
			return
		}
	} else if serverUpdateID > ctrl.updateID+1 {
		logs.Info("SyncController:: Sequential sync")
		getUpdateDifference(ctrl, serverUpdateID+1) // +1 cuz in here we dont have serverUpdateID itself too
	}
}
func updateSyncStatus(ctrl *Controller, newStatus domain.SyncStatus) {
	if ctrl.syncStatus == newStatus {
		return
	}
	logs.Info("Sync Controller:: Status Updated",
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
func getContacts(waitGroup *sync.WaitGroup, ctrl *Controller) {
	req := new(msg.ContactsGet)
	reqBytes, _ := req.Marshal()
	ctrl.queueCtrl.ExecuteCommand(
		uint64(domain.SequentialUniqueID()),
		msg.C_ContactsGet,
		reqBytes,
		func() {
			getContacts(waitGroup, ctrl)
		},
		func(m *msg.MessageEnvelope) {
			waitGroup.Done()
			// Controller applier will take care of this
		},
		false,
	)
}
func getAllDialogs(waitGroup *sync.WaitGroup, ctrl *Controller, offset int32, limit int32) {
	logs.Info("SyncController:: getAllDialogs",
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
			// If timeout, then retry the request
			logs.Warn("getAllDialogs() -> onTimeout() retry to getAllDialogs()")
			getAllDialogs(waitGroup, ctrl, offset, limit)
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
					mMessages[message.ID] = message
				}

				for _, dialog := range x.Dialogs {
					topMessage, _ := mMessages[dialog.TopMessageID]
					if topMessage == nil {
						logs.Error("getAllDialogs() -> onSuccessCallback() -> dialog TopMessage is null",
							zap.Int64("MessageID", dialog.TopMessageID),
						)
						continue
					}
					messageHole.InsertFill(dialog.PeerID, dialog.PeerType, dialog.TopMessageID, dialog.TopMessageID)
				}

				repo.Users.Save(x.Users...)
				repo.Groups.Save(x.Groups...)
				repo.Messages.Save(x.Messages...)

				if x.Count > offset+limit {
					getAllDialogs(waitGroup, ctrl, offset+limit, limit)
				} else {
					waitGroup.Done()
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

	logs.Info("SyncController::getUpdateDifference()",
		zap.Int64("ServerUpdateID", serverUpdateID),
		zap.Int64("ClientUpdateID", ctrl.updateID),
	)

	waitGroup := new(sync.WaitGroup)
	moreTries := 10
	fromUpdateID := ctrl.updateID
	for serverUpdateID > ctrl.updateID {
		if fromUpdateID == ctrl.updateID {
			if moreTries--; moreTries < 0 {
				break
			}
		}
		fromUpdateID = ctrl.updateID
		limit := serverUpdateID - ctrl.updateID
		if limit > 100 {
			limit = 100
		}
		if limit <= 0 {
			break
		}

		req := new(msg.UpdateGetDifference)
		req.Limit = int32(limit)
		req.From = ctrl.updateID + 1 // +1 cuz we already have ctrl.updateID itself
		reqBytes, _ := req.Marshal()
		_ = ctrl.queueCtrl.ExecuteRealtimeCommand(
			uint64(domain.SequentialUniqueID()),
			msg.C_UpdateGetDifference,
			reqBytes,
			func() {
				logs.Warn("SyncController::getUpdateDifference() -> ExecuteRealtimeCommand() Timeout")
				time.Sleep(time.Second)
			},
			func(m *msg.MessageEnvelope) {
				switch m.Constructor {
				case msg.C_UpdateDifference:
					x := new(msg.UpdateDifference)
					err := x.Unmarshal(m.Message)
					if err != nil {
						logs.Error("onGetDifferenceSucceed()-> Unmarshal()", zap.Error(err))
						return
					}
					waitGroup.Wait() // We wait here, because we DON'T want to process update batches in parallel,
					// just we go to pre-fetch the next batch from the server if any
					if x.MaxUpdateID > ctrl.updateID {
						ctrl.updateID = x.MaxUpdateID
					}
					waitGroup.Add(1)
					go func() {
						onGetDifferenceSucceed(ctrl, x)
						// save UpdateID to DB
						err := repo.System.SaveInt(domain.SkUpdateID, uint64(ctrl.updateID))
						if err != nil {
							logs.Error("onGetDifferenceSucceed()-> SaveInt()", zap.Error(err))
						}
						waitGroup.Done()
					}()
				case msg.C_Error:
					logs.Debug("onGetDifferenceSucceed()-> C_Error",
						zap.String("Error", domain.ParseServerError(m.Message).Error()),
					)
				}

			},
			true,
			false,
		)
	}
	waitGroup.Wait()
}
func onGetDifferenceSucceed(ctrl *Controller, x *msg.UpdateDifference) {
	updContainer := new(msg.UpdateContainer)
	updContainer.Updates = make([]*msg.UpdateEnvelope, 0)
	updContainer.Users = x.Users
	updContainer.Groups = x.Groups
	updContainer.MaxUpdateID = x.MaxUpdateID
	updContainer.MinUpdateID = x.MinUpdateID

	logs.Info("SyncController:: onGetDifferenceSucceed",
		zap.Int64("MaxUpdateID", x.MaxUpdateID),
		zap.Int64("MinUpdateID", x.MinUpdateID),
	)

	// save Groups & Users
	repo.Groups.Save(x.Groups...)
	repo.Users.Save(x.Users...)

	for _, update := range x.Updates {
		if applier, ok := ctrl.updateAppliers[update.Constructor]; ok {
			externalHandlerUpdates := applier(update)
			updContainer.Updates = append(updContainer.Updates, externalHandlerUpdates...)
		}
	}
	updContainer.Length = int32(len(updContainer.Updates))

	// We wait here, if any unfinished parallel job has not been finished yet
	ctrl.waitGroup.Wait()

	// wrapped to UpdateContainer
	buff, _ := updContainer.Marshal()
	uiexec.Ctx().Exec(func() {
		if ctrl.onUpdateMainDelegate != nil {
			ctrl.onUpdateMainDelegate(msg.C_UpdateContainer, buff)
		}
	})

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

// Start controller
func (ctrl *Controller) Start() {
	logs.Info("SyncController::  Start")

	// Load the latest UpdateID stored in DB
	if v, err := repo.System.LoadInt(domain.SkUpdateID); err != nil {
		err := repo.System.SaveInt(domain.SkUpdateID, 0)
		if err != nil {
			logs.Error("Start()-> SaveInt()", zap.Error(err))
		}
		ctrl.updateID = 0
	} else {
		ctrl.updateID = int64(v)
	}

	// set default value to synced status
	updateSyncStatus(ctrl, domain.Synced)

	go ctrl.watchDog()
}

// Stop controller
func (ctrl *Controller) Stop() {
	logs.Debug("StopServices-SyncController::Stop() -> Called")
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
		switch m.Constructor {
		case msg.C_MessagesSent:
		// Do nothing
		default:
			if applier, ok := ctrl.messageAppliers[m.Constructor]; ok {
				applier(m)
			}
		}
	}
}

// UpdateHandler receives update to cache them in client DB
func (ctrl *Controller) UpdateHandler(updateContainer *msg.UpdateContainer) {
	logs.Debug("SyncController::UpdateHandler() Called",
		zap.Int64("ctrl.UpdateID", ctrl.updateID),
		zap.Int64("MaxID", updateContainer.MaxUpdateID),
		zap.Int64("MinID", updateContainer.MinUpdateID),
		zap.Int("Count", len(updateContainer.Updates)),
	)

	// Check if update has been already applied
	if updateContainer.MinUpdateID != 0 && ctrl.updateID >= updateContainer.MinUpdateID {
		return
	}

	ctrl.lastUpdateReceived = time.Now()

	// Check if we are out of sync with server, if yes, then call the sync() function
	// We call it in blocking mode,
	if ctrl.updateID < updateContainer.MinUpdateID-1 {
		go ctrl.sync()
		return
	}

	udpContainer := new(msg.UpdateContainer)
	udpContainer.Updates = make([]*msg.UpdateEnvelope, 0)
	udpContainer.MaxUpdateID = updateContainer.MaxUpdateID
	udpContainer.MinUpdateID = updateContainer.MinUpdateID
	udpContainer.Users = updateContainer.Users
	udpContainer.Groups = updateContainer.Groups

	// save Groups & Users
	repo.Groups.Save(updateContainer.Groups...)
	repo.Users.Save(updateContainer.Users...)

	for _, update := range updateContainer.Updates {
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

	// We wait here, if any unfinished parallel job has not been finished yet
	ctrl.waitGroup.Wait()

	// save updateID after processing messages
	if ctrl.updateID < updateContainer.MaxUpdateID {
		ctrl.updateID = updateContainer.MaxUpdateID
		err := repo.System.SaveInt(domain.SkUpdateID, uint64(ctrl.updateID))
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
	return
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
	contactsGetHash, _ := repo.System.LoadInt(domain.SkContactsGetHash)
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

func (ctrl *Controller) UpdateSalt() {
	for !salt.UpdateSalt() {
		ctrl.getServerSalt()
	}
}
func (ctrl *Controller) getServerSalt() {
	serverSaltReq := new(msg.SystemGetSalts)
	serverSaltReqBytes, _ := serverSaltReq.Marshal()

	for {
		err := ctrl.queueCtrl.ExecuteRealtimeCommand(
			uint64(domain.SequentialUniqueID()),
			msg.C_SystemGetSalts,
			serverSaltReqBytes,
			nil,
			func(m *msg.MessageEnvelope) {
				switch m.Constructor {
				case msg.C_SystemSalts:
					s := new(msg.SystemSalts)
					err := s.Unmarshal(m.Message)
					if err != nil {
						logs.Error("Salt:: Error On Unmarshal Server Message, C_SystemSalts", zap.Error(err))
						return
					}

					var saltArray []domain.Slt
					for idx, saltValue := range s.Salts {
						slt := domain.Slt{}
						slt.Timestamp = s.StartsFrom + (s.Duration/int64(time.Second))*int64(idx)
						slt.Value = saltValue
						saltArray = append(saltArray, slt)
					}
					b, _ := json.Marshal(saltArray)
					err = repo.System.SaveString(domain.SkSystemSalts, string(b))
					if err != nil {
						logs.Error("Salt:: save To DB", zap.Error(err))
					}
				case msg.C_Error:
					e := new(msg.Error)
					_ = m.Unmarshal(m.Message)
					logs.Error("Salt:: Error Response from Server",
						zap.String("Code", e.Code),
						zap.String("Item", e.Items),
					)
				}
			},
			true,
			false,
		)
		if err == nil {
			break
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}
