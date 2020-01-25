package syncCtrl

import (
	"git.ronaksoftware.com/ronak/riversdk/msg/chat"
	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_file"
	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_network"
	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_queue"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"git.ronaksoftware.com/ronak/riversdk/pkg/uiexec"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"go.uber.org/zap"
	"sort"
	"sync"
	"time"
)

// Config sync controller required configs
type Config struct {
	ConnInfo           domain.RiverConfigurator
	NetworkCtrl        *networkCtrl.Controller
	QueueCtrl          *queueCtrl.Controller
	FileCtrl           *fileCtrl.Controller
	SyncStatusChangeCB domain.SyncStatusChangeCallback
	UpdateReceivedCB   domain.UpdateReceivedCallback
	AppUpdateCB        domain.AppUpdateCallback
}

// Controller cache received data from server to client DB
type Controller struct {
	connInfo    domain.RiverConfigurator
	networkCtrl *networkCtrl.Controller
	queueCtrl   *queueCtrl.Controller
	fileCtrl    *fileCtrl.Controller

	syncStatus         domain.SyncStatus
	lastUpdateReceived time.Time
	updateID           int64
	updateAppliers     map[int64]domain.UpdateApplier
	messageAppliers    map[int64]domain.MessageApplier
	stopChannel        chan bool
	userID             int64

	// Callbacks
	syncStatusChangeCallback domain.SyncStatusChangeCallback
	updateReceivedCallback   domain.UpdateReceivedCallback
	appUpdateCallback        domain.AppUpdateCallback
}

// NewSyncController create new instance
func NewSyncController(config Config) *Controller {
	ctrl := new(Controller)
	ctrl.stopChannel = make(chan bool)
	ctrl.connInfo = config.ConnInfo
	ctrl.queueCtrl = config.QueueCtrl
	ctrl.networkCtrl = config.NetworkCtrl
	ctrl.fileCtrl = config.FileCtrl

	if config.UpdateReceivedCB == nil {
		config.UpdateReceivedCB = func(constructor int64, msg []byte) {}
	}
	ctrl.updateReceivedCallback = config.UpdateReceivedCB

	if config.SyncStatusChangeCB == nil {
		config.SyncStatusChangeCB = func(newStatus domain.SyncStatus) {}
	}
	ctrl.syncStatusChangeCallback = config.SyncStatusChangeCB

	if config.AppUpdateCB == nil {
		config.AppUpdateCB = func(version string, updateAvailable, force bool) {}
	}
	ctrl.appUpdateCallback = config.AppUpdateCB

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
		msg.C_UpdateGroupAdmins:           ctrl.updateGroupAdmins,
		msg.C_UpdateDialogPinned:          ctrl.updateDialogPinned,
		msg.C_UpdateAccountPrivacy:        ctrl.updateAccountPrivacy,
		msg.C_UpdateDraftMessage:          ctrl.updateDraftMessage,
		msg.C_UpdateDraftMessageCleared:   ctrl.updateDraftMessageCleared,
		msg.C_UpdateLabelSet:              ctrl.updateLabelSet,
		msg.C_UpdateLabelDeleted:          ctrl.updateLabelDeleted,
		msg.C_UpdateLabelItemsAdded:       ctrl.updateLabelItemsAdded,
		msg.C_UpdateLabelItemsRemoved:     ctrl.updateLabelItemsRemoved,
		msg.C_UpdateUserBlocked:           ctrl.updateUserBlocked,
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
		msg.C_LabelsMany:        ctrl.labelsMany,
		msg.C_LabelItems:        ctrl.labelItems,
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
			ctrl.Sync()
		case <-ctrl.stopChannel:
			logs.Info("SyncCtrl's watchDog Stopped")
			return
		}
	}
}

func (ctrl *Controller) SetSynced() {
	updateSyncStatus(ctrl, domain.Synced)
}

func (ctrl *Controller) Sync() {
	_, _, _ = domain.SingleFlight.Do("Sync", func() (i interface{}, e error) {
		// There is no need to sync when no user has been authorized
		if ctrl.userID == 0 {
			logs.Debug("SyncCtrl does not sync when no user is set")
			return
		}

		// get updateID from server
		var serverUpdateID int64
		var err error
		var maxTry = 3
		for {
			serverUpdateID, err = ctrl.AuthRecall("Sync")
			if err != nil {
				logs.Warn("SyncCtrl got err on AuthRecall", zap.Error(err))
				time.Sleep(time.Duration(ronak.RandomInt64(1000)) * time.Millisecond)
				if maxTry--; maxTry < 0 {
					return
				}
			} else {
				break
			}
		}

		if ctrl.updateID == serverUpdateID {
			updateSyncStatus(ctrl, domain.Synced)
			return
		}

		// Update the sync controller status
		updateSyncStatus(ctrl, domain.Syncing)
		defer updateSyncStatus(ctrl, domain.Synced)

		if ctrl.updateID == 0 || (serverUpdateID-ctrl.updateID) > domain.SnapshotSyncThreshold {
			logs.Info("SyncCtrl goes for a Snapshot sync")

			// Get Contacts from the server
			waitGroup := &sync.WaitGroup{}
			waitGroup.Add(2)
			go ctrl.GetContacts(waitGroup)
			go ctrl.GetAllDialogs(waitGroup, 0, 50)
			waitGroup.Wait()

			ctrl.updateID = serverUpdateID
			err = repo.System.SaveInt(domain.SkUpdateID, uint64(ctrl.updateID))
			if err != nil {
				logs.Error("SyncCtrl couldn't save the current UpdateID", zap.Error(err))
				return
			}
		} else if serverUpdateID > ctrl.updateID+1 {
			logs.Info("SyncCtrl goes for a Sequential sync")
			getUpdateDifference(ctrl, serverUpdateID)
		}
		return nil, nil
	})
}
func updateUI(ctrl *Controller, dialogs, contacts bool) {
	update := new(msg.ClientUpdateSynced)
	update.Dialogs = dialogs
	update.Contacts = contacts
	bytes, _ := update.Marshal()

	updateEnvelope := new(msg.UpdateEnvelope)
	updateEnvelope.Constructor = msg.C_ClientUpdateSynced
	updateEnvelope.Update = bytes
	updateEnvelope.UpdateID = 0
	updateEnvelope.Timestamp = time.Now().Unix()
	buff, _ := updateEnvelope.Marshal()

	// call external handler
	uiexec.Ctx().Exec(func() {
		ctrl.updateReceivedCallback(msg.C_UpdateEnvelope, buff)
	})
}
func updateSyncStatus(ctrl *Controller, newStatus domain.SyncStatus) {
	if ctrl.syncStatus == newStatus {
		return
	}
	logs.Info("Sync Controller status changed",
		zap.String("Status", newStatus.ToString()),
	)
	ctrl.syncStatus = newStatus
	ctrl.syncStatusChangeCallback(newStatus)
}
func getUpdateDifference(ctrl *Controller, serverUpdateID int64) {
	logs.Info("SyncCtrl calls getUpdateDifference",
		zap.Int64("ServerUpdateID", serverUpdateID),
		zap.Int64("ClientUpdateID", ctrl.updateID),
	)

	waitGroup := sync.WaitGroup{}
	for serverUpdateID > ctrl.updateID {
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
		waitGroup.Add(1)
		ctrl.queueCtrl.RealtimeCommand(
			uint64(domain.SequentialUniqueID()),
			msg.C_UpdateGetDifference,
			reqBytes,
			func() {
				waitGroup.Done()
				logs.Warn("SyncCtrl got timeout on UpdateGetDifference")
			},
			func(m *msg.MessageEnvelope) {
				defer waitGroup.Done()
				switch m.Constructor {
				case msg.C_UpdateDifference:
					x := new(msg.UpdateDifference)
					err := x.Unmarshal(m.Message)
					if err != nil {
						logs.Error("SyncCtrl couldn't unmarshal response (UpdateDifference)", zap.Error(err))
						return
					}
					sort.Slice(x.Updates, func(i, j int) bool {
						return x.Updates[i].UpdateID < x.Updates[j].UpdateID
					})
					onGetDifferenceSucceed(ctrl, x)
					if x.CurrentUpdateID != 0 {
						serverUpdateID = x.CurrentUpdateID
					}

					// If there is no more update then set ClientUpdateID to the ServerUpdateID
					if !x.More {
						ctrl.updateID = x.CurrentUpdateID
					}

					logs.Info("SyncCtrl received UpdateDifference",
						zap.Bool("More", x.More),
						zap.Int64("MinUpdateID", x.MinUpdateID),
						zap.Int64("MaxUpdateID", x.MaxUpdateID),
					)

					// save UpdateID to DB
					err = repo.System.SaveInt(domain.SkUpdateID, uint64(ctrl.updateID))
					if err != nil {
						logs.Error("SyncCtrl couldn't save current UpdateID", zap.Error(err))
					}
				case msg.C_Error:
					logs.Debug("SyncCtrl got error response",
						zap.String("Error", domain.ParseServerError(m.Message).Error()),
					)
				}

			},
			true,
			false,
		)
		waitGroup.Wait()
	}
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
		zap.Int("Length", len(x.Updates)),
	)

	if len(x.Updates) == 0 {
		return
	}

	// save Groups & Users
	repo.Groups.Save(x.Groups...)
	repo.Users.Save(x.Users...)

	for _, update := range x.Updates {
		if applier, ok := ctrl.updateAppliers[update.Constructor]; ok {
			externalHandlerUpdates, err := applier(update)
			if err != nil {
				logs.Warn("Error On UpdateDiff",
					zap.Error(err),
					zap.Int64("UpdateID", update.UpdateID),
					zap.String("Constructor", msg.ConstructorNames[update.Constructor]),
				)
				return
			}
			updContainer.Updates = append(updContainer.Updates, externalHandlerUpdates...)
		} else {
			updContainer.Updates = append(updContainer.Updates, update)
		}
		if update.UpdateID != 0 {
			ctrl.updateID = update.UpdateID
		}
	}
	updContainer.Length = int32(len(updContainer.Updates))

	// wrapped to UpdateContainer
	buff, _ := updContainer.Marshal()
	uiexec.Ctx().Exec(func() {
		ctrl.updateReceivedCallback(msg.C_UpdateContainer, buff)
	})

}

func (ctrl *Controller) SetUserID(userID int64) {
	ctrl.userID = userID
	logs.Debug("SyncCtrl user is set",
		zap.Int64("userID", userID),
	)
}

func (ctrl *Controller) GetUserID() int64 {
	return ctrl.userID
}

// Start controller
func (ctrl *Controller) Start() {
	logs.Info("SyncCtrl started")

	// Load the latest UpdateID stored in DB
	if v, err := repo.System.LoadInt(domain.SkUpdateID); err != nil {
		err := repo.System.SaveInt(domain.SkUpdateID, 0)
		if err != nil {
			logs.Error("SyncCtrl couldn't save current UpdateID", zap.Error(err))
		}
		ctrl.updateID = 0
	} else {
		ctrl.updateID = int64(v)
	}

	// set default value to synced status
	updateSyncStatus(ctrl, domain.OutOfSync)

	go ctrl.watchDog()
}

// Stop controller
func (ctrl *Controller) Stop() {
	logs.Debug("SyncCtrl calls stop")
	ctrl.stopChannel <- true // for watchDog()
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
func (ctrl *Controller) UpdateHandler(updateContainer *msg.UpdateContainer, outOfSync bool) {
	logs.Debug("SyncCtrl calls UpdateHandler",
		zap.Int64("ctrl.UpdateID", ctrl.updateID),
		zap.Int64("MaxID", updateContainer.MaxUpdateID),
		zap.Int64("MinID", updateContainer.MinUpdateID),
		zap.Int("Count", len(updateContainer.Updates)),
	)

	ctrl.lastUpdateReceived = time.Now()

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
		if outOfSync && update.UpdateID != 0 {
			continue
		}
		applier, ok := ctrl.updateAppliers[update.Constructor]
		if ok {
			externalHandlerUpdates, err := applier(update)
			if err != nil {
				logs.Error("SyncCtrl got error on update applier", zap.Error(err))
				return
			}
			logs.Debug("SyncCtrl applied update", zap.String("Constructor", msg.ConstructorNames[update.Constructor]))
			switch update.Constructor {
			case msg.C_UpdateMessageID:
			default:
				udpContainer.Updates = append(udpContainer.Updates, externalHandlerUpdates...)
			}
		} else {
			udpContainer.Updates = append(udpContainer.Updates, update)
		}
		if update.UpdateID != 0 {
			ctrl.updateID = update.UpdateID
		}
	}

	err := repo.System.SaveInt(domain.SkUpdateID, uint64(ctrl.updateID))
	if err != nil {
		logs.Error("SyncCtrl got error on save UpdateID", zap.Error(err))
	}

	udpContainer.Length = int32(len(udpContainer.Updates))

	// wrapped to UpdateContainer
	buff, _ := udpContainer.Marshal()
	uiexec.Ctx().Exec(func() {
		ctrl.updateReceivedCallback(msg.C_UpdateContainer, buff)
	})

	return
}

// UpdateID returns current updateID
func (ctrl *Controller) UpdateID() int64 {
	return ctrl.updateID
}

// ClearUpdateID reset updateID
func (ctrl *Controller) ResetIDs() {
	ctrl.updateID = 0
	ctrl.userID = 0
}

// ContactImportFromServer import contact from server
func (ctrl *Controller) ContactImportFromServer() {
	contactsGetHash, _ := repo.System.LoadInt(domain.SkContactsGetHash)
	contactGetReq := &msg.ContactsGet{
		Crc32Hash: uint32(contactsGetHash),
	}
	contactGetBytes, _ := contactGetReq.Marshal()
	ctrl.queueCtrl.RealtimeCommand(
		uint64(domain.SequentialUniqueID()),
		msg.C_ContactsGet, contactGetBytes,
		nil, nil, false, false,
	)
	logs.Debug("SyncCtrl call ContactsGet", zap.Uint32("Hash", contactGetReq.Crc32Hash))
}

// GetSyncStatus
func (ctrl *Controller) GetSyncStatus() domain.SyncStatus {
	return ctrl.syncStatus
}
