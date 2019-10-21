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
	ronak "git.ronaksoftware.com/ronak/toolbox"
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
	syncLock int32
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
			ctrl.Sync()
		case <-ctrl.stopChannel:
			logs.Info("SyncCtrl's watchDog Stopped")
			return
		}
	}
}

func (ctrl *Controller) Sync() {
	// Check if sync function is already running, then return otherwise lock it and continue
	if !atomic.CompareAndSwapInt32(&ctrl.syncLock, 0, 1) {
		logs.Debug("SyncCtrl is already syncing ...")
		return
	}
	defer atomic.StoreInt32(&ctrl.syncLock, 0)

	// There is no need to sync when no user has been authorized
	if ctrl.userID == 0 {
		logs.Debug("SyncCtrl does not sync when no user is set")
		return
	}

	// get updateID from server
	var serverUpdateID int64
	var err error
	for {
		serverUpdateID, err = getUpdateState(ctrl)
		if err != nil {
			logs.Warn("SyncCtrl got err on GetUpdateState", zap.Error(err))
			time.Sleep(time.Duration(ronak.RandomInt64(1000)) * time.Millisecond)
		} else {
			break
		}
	}

	if ctrl.updateID == serverUpdateID {
		return
	}

	// Update the sync controller status
	updateSyncStatus(ctrl, domain.Syncing)
	defer updateSyncStatus(ctrl, domain.Synced)

	if ctrl.updateID == 0 || (serverUpdateID-ctrl.updateID) > domain.SnapshotSyncThreshold {
		logs.Info("SyncCtrl goes for a Snapshot sync")

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
			logs.Error("SyncCtrl couldn't save the current UpdateID", zap.Error(err))
			return
		}
	} else if serverUpdateID > ctrl.updateID+1 {
		logs.Info("SyncCtrl goes for a Sequential sync")
		getUpdateDifference(ctrl, serverUpdateID)
	}
}
func updateSyncStatus(ctrl *Controller, newStatus domain.SyncStatus) {
	if ctrl.syncStatus == newStatus {
		return
	}
	logs.Info("Sync Controller status changed",
		zap.String("Status", newStatus.ToString()),
	)
	ctrl.syncStatus = newStatus

	if ctrl.onSyncStatusChange != nil {
		ctrl.onSyncStatusChange(newStatus)
	}
}
func getUpdateState(ctrl *Controller) (updateID int64, err error) {
	logs.Debug("SyncCtrl calls getUpdateState")
	updateID = 0
	if !ctrl.networkCtrl.Connected() {
		return -1, domain.ErrNoConnection
	}

	req := new(msg.UpdateGetState)
	reqBytes, _ := req.Marshal()

	// this waitGroup is required cuz our callbacks will be called in UIExecutor go routine
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(1)
	ctrl.queueCtrl.RealtimeCommand(
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
	logs.Debug("SyncCtrl calls getContacts")
	req := new(msg.ContactsGet)
	reqBytes, _ := req.Marshal()
	ctrl.queueCtrl.EnqueueCommand(
		uint64(domain.SequentialUniqueID()),
		msg.C_ContactsGet,
		reqBytes,
		func() {
			getContacts(waitGroup, ctrl)
		},
		func(m *msg.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_Error:
				errMsg := new(msg.Error)
				_ = errMsg.Unmarshal(m.Message)
				getContacts(waitGroup, ctrl)
			default:
				waitGroup.Done()
			}
			// Controller applier will take care of this
		},
		false,
	)
}
func getAllDialogs(waitGroup *sync.WaitGroup, ctrl *Controller, offset int32, limit int32) {
	logs.Info("SyncCtrl calls getAllDialogs",
		zap.Int32("Offset", offset),
		zap.Int32("Limit", limit),
	)
	req := new(msg.MessagesGetDialogs)
	req.Limit = limit
	req.Offset = offset
	reqBytes, _ := req.Marshal()
	ctrl.queueCtrl.EnqueueCommand(
		uint64(domain.SequentialUniqueID()),
		msg.C_MessagesGetDialogs,
		reqBytes,
		func() {
			// If timeout, then retry the request
			logs.Warn("Timeout! on GetAllDialogs, retrying ...")
			getAllDialogs(waitGroup, ctrl, offset, limit)
		},
		func(m *msg.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_Error:
				logs.Error("SyncCtrl got error response on MessagesGetDialogs",
					zap.String("Error", domain.ParseServerError(m.Message).Error()),
				)
				getAllDialogs(waitGroup, ctrl, offset, limit)
			case msg.C_MessagesDialogs:
				x := new(msg.MessagesDialogs)
				err := x.Unmarshal(m.Message)
				if err != nil {
					logs.Error("SyncCtrl cannot unmarshal server response on MessagesGetDialogs", zap.Error(err))
					return
				}
				mMessages := make(map[int64]*msg.UserMessage)
				for _, message := range x.Messages {
					mMessages[message.ID] = message
				}

				for _, dialog := range x.Dialogs {
					topMessage, _ := mMessages[dialog.TopMessageID]
					if topMessage == nil {
						logs.Error("SyncCtrl received a dialog with no top message",
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
			}
		},
		false,
	)
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
					onGetDifferenceSucceed(ctrl, x)

					// If there is no more update then set ClientUpdateID to the ServerUpdateID
					if !x.More {
						ctrl.updateID = x.CurrentUpdateID
					}

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
			if update.UpdateID != 0 {
				ctrl.updateID = update.UpdateID
			}
			updContainer.Updates = append(updContainer.Updates, externalHandlerUpdates...)
		}
	}
	updContainer.Length = int32(len(updContainer.Updates))

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
	updateSyncStatus(ctrl, domain.Synced)

	go ctrl.watchDog()
}

// Stop controller
func (ctrl *Controller) Stop() {
	logs.Debug("SyncCtrl calls stop")
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
			if update.UpdateID != 0 {
				ctrl.updateID = update.UpdateID
			}
			switch update.Constructor {
			case msg.C_UpdateMessageID:
			default:
				udpContainer.Updates = append(udpContainer.Updates, externalHandlerUpdates...)
			}
		} else {
			udpContainer.Updates = append(udpContainer.Updates, update)
		}
	}

	err := repo.System.SaveInt(domain.SkUpdateID, uint64(ctrl.updateID))
	if err != nil {
		logs.Error("SyncCtrl got error on save UpdateID", zap.Error(err))
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
func (ctrl *Controller) ResetIDs() {
	ctrl.updateID = 0
	ctrl.userID = 0
}

// ContactImportFromServer import contact from server
func (ctrl *Controller) ContactImportFromServer() {
	contactsGetHash, _ := repo.System.LoadInt(domain.SkContactsGetHash)
	contactGetReq := new(msg.ContactsGet)
	contactGetReq.Crc32Hash = uint32(contactsGetHash)
	contactGetBytes, _ := contactGetReq.Marshal()
	ctrl.queueCtrl.RealtimeCommand(
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
		ctrl.SendGetServerSalt()
	}
}

func (ctrl *Controller) SendGetServerSalt() {
	serverSaltReq := new(msg.SystemGetSalts)
	serverSaltReqBytes, _ := serverSaltReq.Marshal()

	keepGoing := true
	for keepGoing {
		if ctrl.networkCtrl.GetQuality() == domain.NetworkDisconnected {
			return
		}
		ctrl.queueCtrl.RealtimeCommand(
			uint64(domain.SequentialUniqueID()),
			msg.C_SystemGetSalts,
			serverSaltReqBytes,
			func() {
				time.Sleep(time.Second)
			},
			func(m *msg.MessageEnvelope) {
				switch m.Constructor {
				case msg.C_SystemSalts:
					s := new(msg.SystemSalts)
					err := s.Unmarshal(m.Message)
					if err != nil {
						logs.Error("SyncCtrl couldn't unmarshal SystemSalts", zap.Error(err))
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
						logs.Error("SyncCtrl couldn't save SystemSalts in the db", zap.Error(err))
						return
					}
					keepGoing = false
				case msg.C_Error:
					e := new(msg.Error)
					_ = m.Unmarshal(m.Message)
					logs.Error("SyncCtrl received error response for SystemGetSalts (Error)",
						zap.String("Code", e.Code),
						zap.String("Item", e.Items),
					)
					time.Sleep(time.Second)
				}
			},
			true,
			false,
		)

	}
}

func (ctrl *Controller) SendAuthRecall(waitGroup *sync.WaitGroup) {
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		ctrl.UpdateSalt()
		req := msg.AuthRecall{}
		reqBytes, _ := req.Marshal()

		// send auth recall until it succeed
		keepGoing := true
		for keepGoing {
			if ctrl.networkCtrl.GetQuality() == domain.NetworkDisconnected {
				return
			}
			// this is priority command that should not passed to queue
			// after auth recall answer got back the queue should send its requests in order to get related updates
			ctrl.queueCtrl.RealtimeCommand(
				uint64(domain.SequentialUniqueID()),
				msg.C_AuthRecall,
				reqBytes,
				func() {
					time.Sleep(time.Duration(ronak.RandomInt(1000)) * time.Millisecond)
				},
				func(m *msg.MessageEnvelope) {
					if m.Constructor == msg.C_AuthRecalled {
						x := new(msg.AuthRecalled)
						err := x.Unmarshal(m.Message)
						if err != nil {
							logs.Error("We couldn't unmarshal AuthRecall (AuthRecalled) response", zap.Error(err))
							return
						}
						keepGoing = false
					}
				},
				true,
				false,
			)
		}

	}()
}

func (ctrl *Controller) SendGetServerTime(waitGroup *sync.WaitGroup) {
	waitGroup.Add(1)
	go func() {
		timeReq := new(msg.SystemGetServerTime)
		timeReqBytes, _ := timeReq.Marshal()
		defer waitGroup.Done()
		keepGoing := true
		for keepGoing {
			if ctrl.networkCtrl.GetQuality() == domain.NetworkDisconnected {
				return
			}
			ctrl.queueCtrl.RealtimeCommand(
				uint64(domain.SequentialUniqueID()),
				msg.C_SystemGetServerTime,
				timeReqBytes,
				func() {
					time.Sleep(time.Duration(ronak.RandomInt(1000)) * time.Millisecond)
				},
				func(m *msg.MessageEnvelope) {
					switch m.Constructor {
					case msg.C_SystemServerTime:
						x := new(msg.SystemServerTime)
						err := x.Unmarshal(m.Message)
						if err != nil {
							logs.Error("We couldn't unmarshal SystemGetServerTime response", zap.Error(err))
							return
						}
						clientTime := time.Now().Unix()
						serverTime := x.Timestamp
						domain.TimeDelta = time.Duration(serverTime-clientTime) * time.Second

						logs.Debug("SystemServerTime received",
							zap.Int64("ServerTime", serverTime),
							zap.Int64("ClientTime", clientTime),
							zap.Duration("Difference", domain.TimeDelta),
						)
						keepGoing = false
					case msg.C_Error:
						logs.Warn("We received error on GetSystemServerTime", zap.Error(domain.ParseServerError(m.Message)))
						time.Sleep(time.Duration(ronak.RandomInt(1000)) * time.Millisecond)
					}
				},
				true, false,
			)
		}
	}()

}
