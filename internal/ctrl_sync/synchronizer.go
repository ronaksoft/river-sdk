package syncCtrl

import (
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/ctrl_file"
	"git.ronaksoft.com/river/sdk/internal/ctrl_network"
	"git.ronaksoft.com/river/sdk/internal/ctrl_queue"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/uiexec"
	"github.com/gobwas/pool/pbytes"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/registry"
	"github.com/ronaksoft/rony/tools"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
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
	userID             int64

	// Callbacks
	syncStatusChangeCallback domain.SyncStatusChangeCallback
	updateReceivedCallback   domain.UpdateReceivedCallback
	appUpdateCallback        domain.AppUpdateCallback
}

// NewSyncController create new instance
func NewSyncController(config Config) *Controller {
	ctrl := new(Controller)
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
		msg.C_UpdateAccountPrivacy:        ctrl.updateAccountPrivacy,
		msg.C_UpdateDialogPinned:          ctrl.updateDialogPinned,
		msg.C_UpdateDraftMessage:          ctrl.updateDraftMessage,
		msg.C_UpdateDraftMessageCleared:   ctrl.updateDraftMessageCleared,
		msg.C_UpdateGroupAdmins:           ctrl.updateGroupAdmins,
		msg.C_UpdateGroupParticipantAdmin: ctrl.updateGroupParticipantAdmin,
		msg.C_UpdateGroupPhoto:            ctrl.updateGroupPhoto,
		msg.C_UpdateLabelDeleted:          ctrl.updateLabelDeleted,
		msg.C_UpdateLabelItemsAdded:       ctrl.updateLabelItemsAdded,
		msg.C_UpdateLabelItemsRemoved:     ctrl.updateLabelItemsRemoved,
		msg.C_UpdateLabelSet:              ctrl.updateLabelSet,
		msg.C_UpdateMessageEdited:         ctrl.updateMessageEdited,
		msg.C_UpdateMessageID:             ctrl.updateMessageID,
		msg.C_UpdateMessagePinned:         ctrl.updateMessagePinned,
		msg.C_UpdateMessagesDeleted:       ctrl.updateMessagesDeleted,
		msg.C_UpdateNewMessage:            ctrl.updateNewMessage,
		msg.C_UpdateNotifySettings:        ctrl.updateNotifySettings,
		msg.C_UpdateReaction:              ctrl.updateReaction,
		msg.C_UpdateReadHistoryInbox:      ctrl.updateReadHistoryInbox,
		msg.C_UpdateReadHistoryOutbox:     ctrl.updateReadHistoryOutbox,
		msg.C_UpdateReadMessagesContents:  ctrl.updateReadMessagesContents,
		msg.C_UpdateTeam:                  ctrl.updateTeam,
		msg.C_UpdateTeamCreated:           ctrl.updateTeamCreated,
		msg.C_UpdateTeamMemberAdded:       ctrl.updateTeamMemberAdded,
		msg.C_UpdateTeamMemberRemoved:     ctrl.updateTeamMemberRemoved,
		msg.C_UpdateTeamMemberStatus:      ctrl.updateTeamMemberStatus,
		msg.C_UpdateUserBlocked:           ctrl.updateUserBlocked,
		msg.C_UpdateUsername:              ctrl.updateUsername,
		msg.C_UpdateUserPhoto:             ctrl.updateUserPhoto,
	}
	ctrl.messageAppliers = map[int64]domain.MessageApplier{
		msg.C_AuthAuthorization:    ctrl.authAuthorization,
		msg.C_AuthSentCode:         ctrl.authSentCode,
		msg.C_BotResults:           ctrl.botResults,
		msg.C_ContactsImported:     ctrl.contactsImported,
		msg.C_ContactsMany:         ctrl.contactsMany,
		msg.C_ContactsTopPeers:     ctrl.contactsTopPeers,
		msg.C_GroupFull:            ctrl.groupFull,
		msg.C_LabelItems:           ctrl.labelItems,
		msg.C_LabelsMany:           ctrl.labelsMany,
		msg.C_MessagesDialogs:      ctrl.messagesDialogs,
		msg.C_MessagesMany:         ctrl.messagesMany,
		msg.C_MessagesReactionList: ctrl.reactionList,
		msg.C_SavedGifs:            ctrl.savedGifs,
		msg.C_SystemConfig:         ctrl.systemConfig,
		msg.C_TeamMembers:          ctrl.teamMembers,
		msg.C_TeamsMany:            ctrl.teamsMany,
		msg.C_UsersMany:            ctrl.usersMany,
		msg.C_WallPapersMany:       ctrl.wallpapersMany,
	}
	return ctrl
}

// watchDog
// Checks if we have not received any updates since last watch tries to re-sync with server.
func (ctrl *Controller) watchDog() {
	syncTime := 3 * time.Minute
	t := time.NewTimer(syncTime)
	for {
		if !t.Stop() {
			<-t.C
		}
		t.Reset(syncTime)
		select {
		case <-t.C:
			// Skip if we are not connected to server
			if ctrl.networkCtrl.GetQuality() != domain.NetworkConnected {
				break
			}

			now := time.Now()
			// Check if we were not syncing in the last 3 minutes
			if now.Sub(ctrl.lastUpdateReceived) > syncTime {
				go ctrl.Sync()
			}

		}
	}
}

func (ctrl *Controller) SetSynced() {
	updateSyncStatus(ctrl, domain.Synced)
}

func (ctrl *Controller) Sync() {
	_, _, _ = domain.SingleFlight.Do("Sync", func() (i interface{}, e error) {
		// There is no need to sync when no user has been authorized
		if ctrl.GetUserID() == 0 {
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
				time.Sleep(time.Duration(domain.RandomInt(1000)) * time.Millisecond)
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
			waitGroup.Add(8)
			go ctrl.GetContacts(waitGroup, 0, 0)
			go ctrl.GetAllDialogs(waitGroup, 0, 0, 0, 100)
			go ctrl.GetLabels(waitGroup, 0, 0)
			go ctrl.GetAllTopPeers(waitGroup, 0, 0, msg.TopPeerCategory_Users, 0, 100)
			go ctrl.GetAllTopPeers(waitGroup, 0, 0, msg.TopPeerCategory_Groups, 0, 100)
			go ctrl.GetAllTopPeers(waitGroup, 0, 0, msg.TopPeerCategory_Forwards, 0, 100)
			go ctrl.GetAllTopPeers(waitGroup, 0, 0, msg.TopPeerCategory_BotsMessage, 0, 100)
			go ctrl.GetAllTopPeers(waitGroup, 0, 0, msg.TopPeerCategory_BotsInline, 0, 100)
			waitGroup.Wait()

			ctrl.updateID = serverUpdateID
			err = repo.System.SaveInt(domain.SkUpdateID, uint64(ctrl.updateID))
			if err != nil {
				logs.Error("SyncCtrl couldn't save the current UpdateID", zap.Error(err))
				return
			}
		} else if serverUpdateID >= ctrl.updateID+1 {
			logs.Info("SyncCtrl goes for a Sequential sync")
			getUpdateDifference(ctrl, serverUpdateID)
		}
		return nil, nil
	})
}
func forceUpdateUI(ctrl *Controller, dialogs, contacts, gifs bool) {
	update := &msg.ClientUpdateSynced{}
	update.Dialogs = dialogs
	update.Contacts = contacts
	update.Gifs = gifs
	bytes, _ := update.Marshal()

	updateEnvelope := &msg.UpdateEnvelope{}
	updateEnvelope.Constructor = msg.C_ClientUpdateSynced
	updateEnvelope.Update = bytes
	updateEnvelope.UpdateID = 0
	updateEnvelope.Timestamp = time.Now().Unix()

	// call external handler
	uiexec.ExecUpdate(ctrl.updateReceivedCallback, msg.C_UpdateEnvelope, updateEnvelope)
}
func updateSyncStatus(ctrl *Controller, newStatus domain.SyncStatus) {
	if ctrl.syncStatus == newStatus {
		return
	}
	logs.Info("SyncCtrl status changed",
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
		if limit > 250 {
			limit = 250
		}
		if limit <= 0 {
			break
		}

		// Break the loop if userID is set to zero
		if ctrl.GetUserID() == 0 {
			break
		}
		req := &msg.UpdateGetDifference{
			Limit: int32(limit),
			From:  ctrl.updateID + 1, // +1 cuz we already have ctrl.updateID itself,
		}
		reqBytes, _ := req.Marshal()
		waitGroup.Add(1)
		ctrl.networkCtrl.RealtimeCommand(
			&rony.MessageEnvelope{
				Constructor: msg.C_UpdateGetDifference,
				RequestID:   uint64(domain.SequentialUniqueID()),
				Message:     reqBytes,
			},
			func() {
				waitGroup.Done()
				logs.Warn("SyncCtrl got timeout on UpdateGetDifference")
			},
			func(m *rony.MessageEnvelope) {
				defer waitGroup.Done()
				switch m.Constructor {
				case msg.C_UpdateDifference:
					x := new(msg.UpdateDifference)
					err := x.Unmarshal(m.Message)
					if err != nil {
						logs.Error("SyncCtrl couldn't unmarshal response (UpdateDifference)", zap.Error(err))
						time.Sleep(time.Second)
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
				case rony.C_Error:
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
					zap.String("C", registry.ConstructorName(update.Constructor)),
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

	uiexec.ExecUpdate(ctrl.updateReceivedCallback, msg.C_UpdateContainer, updContainer)
}

func (ctrl *Controller) TeamSync(teamID int64, accessHash uint64, forceUpdate bool) {
	teamKey := fmt.Sprintf("%s.%d", domain.SkTeam, teamID)

	if !forceUpdate {
		lastSync, _ := repo.System.LoadInt64(teamKey)
		if lastSync > 0 {
			// we have been already synced
			return
		}
	}

	// Get Contacts from the server
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(8)
	go ctrl.GetContacts(waitGroup, teamID, accessHash)
	go ctrl.GetAllDialogs(waitGroup, teamID, accessHash, 0, 100)
	go ctrl.GetLabels(waitGroup, teamID, accessHash)
	go ctrl.GetAllTopPeers(waitGroup, teamID, accessHash, msg.TopPeerCategory_Users, 0, 100)
	go ctrl.GetAllTopPeers(waitGroup, teamID, accessHash, msg.TopPeerCategory_Groups, 0, 100)
	go ctrl.GetAllTopPeers(waitGroup, teamID, accessHash, msg.TopPeerCategory_Forwards, 0, 100)
	go ctrl.GetAllTopPeers(waitGroup, teamID, accessHash, msg.TopPeerCategory_BotsMessage, 0, 100)
	go ctrl.GetAllTopPeers(waitGroup, teamID, accessHash, msg.TopPeerCategory_BotsInline, 0, 100)
	waitGroup.Wait()

	// if this is the first time we switch to this team, then lets sync with server
	err := repo.System.SaveInt(teamKey, uint64(tools.TimeUnix()))
	logs.WarnOnErr("Team Sync", err)
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
	logs.Info("SyncCtrl calls stop")
	ctrl.ResetIDs()
	logs.Info("SyncCtrl Stopped")
}

// MessageHandler call appliers-> repository and sync data
func (ctrl *Controller) MessageHandler(messages []*rony.MessageEnvelope) {
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
			logs.Debug("SyncCtrl applied update", zap.String("C", registry.ConstructorName(update.Constructor)))
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
	uiexec.ExecUpdate(ctrl.updateReceivedCallback, msg.C_UpdateContainer, updateContainer)
	return
}

// UpdateID returns current updateID
func (ctrl *Controller) UpdateID() int64 {
	return ctrl.updateID
}

// ClearUpdateID reset updateID
func (ctrl *Controller) ResetIDs() {
	ctrl.updateID = 0
	ctrl.SetUserID(0)
}

// ContactsImport
func (ctrl *Controller) ContactsImport(replace bool, successCB domain.MessageHandler, out *rony.MessageEnvelope) {
	var (
		wg               = sync.WaitGroup{}
		limit            = 250
		maxTry           = 10
		keepGoing        = true
		contactsImported = &msg.ContactsImported{}
	)
	if out == nil {
		out = &rony.MessageEnvelope{}
	}

	for keepGoing {
		phoneContacts, _ := repo.Users.GetPhoneContacts(limit)
		if len(phoneContacts) < limit {
			keepGoing = false
		}

		// We have not more phone contacts left
		if len(phoneContacts) == 0 {
			break
		}

		req := &msg.ContactsImport{
			Replace:  replace,
			Contacts: phoneContacts,
		}
		mo := proto.MarshalOptions{UseCachedSize: true}
		reqBytes := pbytes.GetCap(mo.Size(req))
		reqBytes, _ = mo.MarshalAppend(reqBytes, req)

		wg.Add(1)
		ctrl.queueCtrl.EnqueueCommand(
			&rony.MessageEnvelope{
				Constructor: msg.C_ContactsImport,
				RequestID:   uint64(domain.SequentialUniqueID()),
				Message:     reqBytes,
			},
			func() {
				wg.Done()
				logs.Error("SyncCtrl got timeout on ContactsImport")
			},
			func(m *rony.MessageEnvelope) {
				defer wg.Done()
				switch m.Constructor {
				case msg.C_ContactsImported:
					x := &msg.ContactsImported{}
					err := x.Unmarshal(m.Message)
					if err != nil {
						logs.Error("SyncCtrl got error on ContactsImport when unmarshal", zap.Error(err))
						return
					}
					_ = repo.Users.DeletePhoneContact(phoneContacts...)
					contactsImported.Users = append(contactsImported.Users, x.Users...)
					contactsImported.ContactUsers = append(contactsImported.ContactUsers, x.ContactUsers...)
					out.Fill(out.RequestID, msg.C_ContactsImported, contactsImported)
				case rony.C_Error:
					x := &rony.Error{}
					_ = x.Unmarshal(m.Message)
					out.Fill(out.RequestID, rony.C_Error, x)
					switch {
					case x.Code == msg.ErrCodeRateLimit:
						maxTry = 0
					default:
						logs.Warn("SyncCtrl got error response from server, will retry",
							zap.String("Code", x.Code), zap.String("Item", x.Items),
						)
						time.Sleep(time.Second)
					}
					if maxTry--; maxTry < 0 {
						keepGoing = false
					}
				default:
					logs.Error("SyncCtrl expected ContactsImported but we got something else!!!",
						zap.String("C", registry.ConstructorName(m.Constructor)),
					)
					time.Sleep(time.Second)
					if maxTry--; maxTry < 0 {
						keepGoing = false
					}
				}

			},
			false,
		)
		wg.Wait()
		pbytes.Put(reqBytes)
	}
	if successCB != nil && out != nil {
		successCB(out)
	} else {
		forceUpdateUI(ctrl, false, true, false)
	}
	return
}

// GetSyncStatus
func (ctrl *Controller) GetSyncStatus() domain.SyncStatus {
	return ctrl.syncStatus
}
