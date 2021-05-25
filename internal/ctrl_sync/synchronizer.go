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
	"git.ronaksoft.com/river/sdk/internal/request"
	"git.ronaksoft.com/river/sdk/internal/uiexec"
	"github.com/gobwas/pool/pbytes"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/pools"
	"github.com/ronaksoft/rony/registry"
	"github.com/ronaksoft/rony/tools"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"sort"
	"sync"
	"time"
)

var (
	logger *logs.Logger
)

func init() {
	logger = logs.With("SyncCtrl")
}

// Config sync controller required configs
type Config struct {
	ConnInfo           domain.RiverConfigurator
	NetworkCtrl        *networkCtrl.Controller
	QueueCtrl          *queueCtrl.Controller
	FileCtrl           *fileCtrl.Controller
	SyncStatusChangeCB domain.SyncStatusChangeCallback
	AppUpdateCB        domain.AppUpdateCallback
}

// Controller cache received data from server to client DB
type Controller struct {
	connInfo    domain.RiverConfigurator
	networkCtrl *networkCtrl.Controller
	queueCtrl   *queueCtrl.Controller
	fileCtrl    *fileCtrl.Controller

	syncStatus         domain.SyncStatus
	lastUpdateReceived int64
	updateID           int64
	updateAppliers     map[int64]domain.UpdateApplier
	messageAppliers    map[int64]domain.MessageApplier
	userID             int64

	// Callbacks
	syncStatusChangeCallback domain.SyncStatusChangeCallback
	appUpdateCallback        domain.AppUpdateCallback
}

// NewSyncController create new instance
func NewSyncController(config Config) *Controller {
	ctrl := &Controller{
		connInfo:    config.ConnInfo,
		queueCtrl:   config.QueueCtrl,
		networkCtrl: config.NetworkCtrl,
		fileCtrl:    config.FileCtrl,
	}

	if config.SyncStatusChangeCB == nil {
		config.SyncStatusChangeCB = func(newStatus domain.SyncStatus) {}
	}
	ctrl.syncStatusChangeCallback = config.SyncStatusChangeCB

	if config.AppUpdateCB == nil {
		config.AppUpdateCB = func(version string, updateAvailable, force bool) {}
	}
	ctrl.appUpdateCallback = config.AppUpdateCB

	ctrl.updateAppliers = map[int64]domain.UpdateApplier{}
	ctrl.messageAppliers = map[int64]domain.MessageApplier{}
	return ctrl
}

func (ctrl *Controller) RegisterUpdateApplier(constructor int64, ua domain.UpdateApplier) {
	_, ok := ctrl.updateAppliers[constructor]
	if ok {
		panic(fmt.Sprintf("BUG!!::update applier already registered: %s", registry.ConstructorName(constructor)))
	}
	ctrl.updateAppliers[constructor] = ua
}

func (ctrl *Controller) RegisterMessageApplier(constructor int64, ma domain.MessageApplier) {
	_, ok := ctrl.messageAppliers[constructor]
	if ok {
		panic(fmt.Sprintf("BUG!!::update applier already registered: %s", registry.ConstructorName(constructor)))
	}
	ctrl.messageAppliers[constructor] = ma

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
			if !ctrl.networkCtrl.Connected() {
				break
			}

			// Check if we were not syncing in the last 3 minutes
			if time.Duration(tools.NanoTime()-ctrl.lastUpdateReceived) > syncTime {
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
			logger.Debug("does not sync when no user is set")
			return
		}

		// get updateID from server
		var serverUpdateID int64
		var err error
		var maxTry = 3
		for {
			serverUpdateID, err = ctrl.AuthRecall("Sync")
			if err != nil {
				logger.Warn("got err on AuthRecall", zap.Error(err))
				time.Sleep(time.Duration(domain.RandomInt(1000)) * time.Millisecond)
				if maxTry--; maxTry < 0 {
					return
				}
			} else {
				break
			}
		}

		if ctrl.GetUpdateID() == serverUpdateID {
			updateSyncStatus(ctrl, domain.Synced)
			return
		}

		// Update the sync controller status
		updateSyncStatus(ctrl, domain.Syncing)
		defer updateSyncStatus(ctrl, domain.Synced)

		ctrlUpdateID := ctrl.GetUpdateID()
		if ctrlUpdateID == 0 || (serverUpdateID-ctrlUpdateID) > domain.SnapshotSyncThreshold {
			logger.Info("goes for a Snapshot sync")

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

			if err := ctrl.SetUpdateID(serverUpdateID); err != nil {
				logger.Error("couldn't save the current GetUpdateID", zap.Error(err))
				return
			}
		} else if serverUpdateID >= ctrl.GetUpdateID()+1 {
			logger.Info("goes for a Sequential sync")
			getUpdateDifference(ctrl, serverUpdateID)
		}
		return nil, nil
	})
}
func updateSyncStatus(ctrl *Controller, newStatus domain.SyncStatus) {
	if ctrl.syncStatus == newStatus {
		return
	}
	logger.Info("status changed", zap.String("Status", newStatus.ToString()))
	ctrl.syncStatus = newStatus
	ctrl.syncStatusChangeCallback(newStatus)
}
func getUpdateDifference(ctrl *Controller, serverUpdateID int64) {
	logger.Info("calls UpdateGetDifference",
		zap.Int64("ServerUpdateID", serverUpdateID),
		zap.Int64("ClientUpdateID", ctrl.GetUpdateID()),
	)

	waitGroup := sync.WaitGroup{}
	for serverUpdateID > ctrl.GetUpdateID() {
		limit := serverUpdateID - ctrl.GetUpdateID()
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
			From:  ctrl.GetUpdateID() + 1, // +1 cuz we already have ctrl.updateID itself,
		}
		reqBytes, _ := req.Marshal()
		waitGroup.Add(1)
		ctrl.networkCtrl.WebsocketCommandWithTimeout(
			&rony.MessageEnvelope{
				Constructor: msg.C_UpdateGetDifference,
				RequestID:   uint64(domain.SequentialUniqueID()),
				Message:     reqBytes,
			},
			func() {
				waitGroup.Done()
				logger.Warn("got timeout on UpdateGetDifference")
			},
			func(m *rony.MessageEnvelope) {
				defer waitGroup.Done()
				switch m.Constructor {
				case msg.C_UpdateDifference:
					x := new(msg.UpdateDifference)
					err := x.Unmarshal(m.Message)
					if err != nil {
						logger.Error("couldn't unmarshal response (UpdateDifference)", zap.Error(err))
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
						_ = ctrl.SetUpdateID(x.CurrentUpdateID)
					}

				case rony.C_Error:
					logger.Debug("got error response",
						zap.String("Error", domain.ParseServerError(m.Message).Error()),
					)

				}

			},
			false,
			request.SkipFlusher,
			domain.WebsocketRequestTimeoutLong,
		)
		waitGroup.Wait()
	}
}
func getUpdateTargetID(u *msg.UpdateEnvelope) string {
	switch u.Constructor {
	case msg.C_UpdateNewMessage:
		x := &msg.UpdateNewMessage{}
		_ = x.Unmarshal(u.Update)
		return fmt.Sprintf(fmt.Sprintf("NewMessage_%d_%d_%d", x.Message.TeamID, x.Message.PeerID, x.Message.PeerType))
	case msg.C_UpdateDraftMessage:
		x := &msg.UpdateDraftMessage{}
		_ = x.Unmarshal(u.Update)
		return fmt.Sprintf(fmt.Sprintf("DraftMessage_%d_%d_%d", x.Message.TeamID, x.Message.PeerID, x.Message.PeerType))
	case msg.C_UpdateDraftMessageCleared:
		x := &msg.UpdateDraftMessageCleared{}
		_ = x.Unmarshal(u.Update)
		return fmt.Sprintf(fmt.Sprintf("DraftMessage_%d_%d_%d", x.TeamID, x.Peer.ID, x.Peer.Type))
	case msg.C_UpdateReadHistoryInbox:
		x := &msg.UpdateReadHistoryInbox{}
		_ = x.Unmarshal(u.Update)
		return fmt.Sprintf(fmt.Sprintf("ReadHistoryIn_%d_%d_%d", x.TeamID, x.Peer.ID, x.Peer.Type))
	case msg.C_UpdateReadHistoryOutbox:
		x := &msg.UpdateReadHistoryOutbox{}
		_ = x.Unmarshal(u.Update)
		return fmt.Sprintf(fmt.Sprintf("ReadHistoryOut_%d_%d_%d", x.TeamID, x.Peer.ID, x.Peer.Type))
	default:
		return fmt.Sprintf("%d", u.Constructor)
	}
}
func onGetDifferenceSucceed(ctrl *Controller, x *msg.UpdateDifference) {
	mtx := sync.Mutex{}
	updContainer := &msg.UpdateContainer{
		Updates:     make([]*msg.UpdateEnvelope, 0),
		Users:       x.Users,
		Groups:      x.Groups,
		MaxUpdateID: x.MaxUpdateID,
		MinUpdateID: x.MinUpdateID,
	}
	var (
		startTime = tools.NanoTime()
		timeLapse [2]int64
	)

	logger.Info("received UpdateDifference",
		zap.Int64("MaxUpdateID", x.MaxUpdateID),
		zap.Int64("MinUpdateID", x.MinUpdateID),
		zap.Int("Length", len(x.Updates)),
		zap.Bool("More", x.More),
	)
	defer func() {
		endTime := tools.NanoTime()
		logger.Info("applied UpdateDifference",
			zap.Int("Length", len(x.Updates)),
			zap.Duration("Messages", time.Duration(timeLapse[1]-timeLapse[0])),
			zap.Duration("Others", time.Duration(endTime-timeLapse[1])),
			zap.Duration("D", time.Duration(endTime-startTime)),
		)
	}()

	if len(x.Updates) == 0 {
		return
	}

	// save Groups & Users
	waitGroup := pools.AcquireWaitGroup()
	defer pools.ReleaseWaitGroup(waitGroup)

	waitGroup.Add(2)
	go func() {
		_ = repo.Groups.Save(x.Groups...)
		waitGroup.Done()
	}()
	go func() {
		_ = repo.Users.Save(x.Users...)
		waitGroup.Done()
	}()

	applierFlusher := tools.NewFlusherPool(16, 10, func(targetID string, entries []tools.FlushEntry) {
		for _, e := range entries {
			ue := e.Value().(*msg.UpdateEnvelope)
			if applier, ok := ctrl.updateAppliers[ue.Constructor]; ok {
				externalHandlerUpdates, err := applier(ue)
				if err != nil {
					logger.Warn("got error on UpdateDifference",
						zap.Error(err),
						zap.Int64("UpdateID", ue.UpdateID),
						zap.String("C", registry.ConstructorName(ue.Constructor)),
					)
					break
				}
				mtx.Lock()
				updContainer.Updates = append(updContainer.Updates, externalHandlerUpdates...)
				mtx.Unlock()
			} else {
				mtx.Lock()
				updContainer.Updates = append(updContainer.Updates, ue)
				mtx.Unlock()
			}
		}
	})
	waitGroup.Wait()

	// Separate updates into categories based on their constructor
	var queues [2][]*msg.UpdateEnvelope
	for _, update := range x.Updates {
		switch update.Constructor {
		case msg.C_UpdateNewMessage:
			queues[0] = append(queues[0], update)
		default:
			queues[1] = append(queues[1], update)
		}
	}

	// apply updates based on their priority queue
	for idx, updates := range queues {
		timeLapse[idx] = tools.NanoTime()
		for _, ue := range updates {
			logger.Info("UpdateDifference applies",
				zap.Int64("UpdateID", ue.UpdateID),
				zap.String("C", registry.ConstructorName(ue.Constructor)),
			)
			waitGroup.Add(1)
			applierFlusher.Enter(
				getUpdateTargetID(ue),
				tools.NewEntryWithCallback(ue, func() {
					waitGroup.Done()
				}),
			)
		}
		waitGroup.Wait()
	}

	if x.MaxUpdateID != 0 {
		_ = ctrl.SetUpdateID(x.MaxUpdateID)
	}
	updContainer.Length = int32(len(updContainer.Updates))

	uiexec.ExecUpdate(msg.C_UpdateContainer, updContainer)
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
	logger.WarnOnErr("Team Sync", err)
}

func (ctrl *Controller) SetUserID(userID int64) {
	ctrl.userID = userID
	logger.Debug("user is set",
		zap.Int64("userID", userID),
	)
}

func (ctrl *Controller) GetUserID() int64 {
	return ctrl.userID
}

// GetUpdateID returns current updateID
func (ctrl *Controller) GetUpdateID() int64 {
	return ctrl.updateID
}

// SetUpdateID set the controller last update id
func (ctrl *Controller) SetUpdateID(id int64) error {
	ctrl.updateID = id
	return repo.System.SaveInt(domain.SkUpdateID, uint64(id))
}

// Start controller
func (ctrl *Controller) Start() {
	logger.Info("started")

	// Load the latest GetUpdateID stored in DB
	if v, err := repo.System.LoadInt(domain.SkUpdateID); err != nil {
		err := repo.System.SaveInt(domain.SkUpdateID, 0)
		if err != nil {
			logger.Error("couldn't save current GetUpdateID", zap.Error(err))
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
	logger.Info("calls stop")
	ctrl.ResetIDs()
	logger.Info("Stopped")
}

// MessageApplier call appliers-> repository and sync data
func (ctrl *Controller) MessageApplier(messages []*rony.MessageEnvelope) {
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

// UpdateApplier receives update to cache them in client DB
func (ctrl *Controller) UpdateApplier(updateContainer *msg.UpdateContainer, outOfSync bool) {
	ctrl.lastUpdateReceived = tools.NanoTime()

	udpContainer := &msg.UpdateContainer{
		Updates:     make([]*msg.UpdateEnvelope, 0),
		MaxUpdateID: updateContainer.MaxUpdateID,
		MinUpdateID: updateContainer.MinUpdateID,
		Users:       updateContainer.Users,
		Groups:      updateContainer.Groups,
	}

	// save Groups & Users
	waitGroup := pools.AcquireWaitGroup()
	waitGroup.Add(2)
	go func() {
		_ = repo.Groups.Save(updateContainer.Groups...)
		waitGroup.Done()
	}()
	go func() {
		_ = repo.Users.Save(updateContainer.Users...)
		waitGroup.Done()
	}()
	waitGroup.Wait()
	pools.ReleaseWaitGroup(waitGroup)

	logger.Debug("receives UpdateContainer",
		zap.Int64("ctrl.GetUpdateID", ctrl.GetUpdateID()),
		zap.Int64("MaxID", updateContainer.MaxUpdateID),
		zap.Int64("MinID", updateContainer.MinUpdateID),
		zap.Int("Count", len(updateContainer.Updates)),
	)

	for _, update := range updateContainer.Updates {

		if outOfSync && update.UpdateID != 0 {
			continue
		}
		applier, ok := ctrl.updateAppliers[update.Constructor]
		if ok {
			logger.Debug("applies Update",
				zap.Int64("ctrl.GetUpdateID", ctrl.GetUpdateID()),
				zap.Int64("MaxID", updateContainer.MaxUpdateID),
				zap.Int64("MinID", updateContainer.MinUpdateID),
				zap.String("C", registry.ConstructorName(update.Constructor)),
			)

			externalHandlerUpdates, err := applier(update)
			if err != nil {
				logger.Error("got error on update applier",
					zap.Error(err),
					zap.String("C", registry.ConstructorName(update.Constructor)),
					zap.Int64("UpdateID", update.UpdateID),
				)
				return
			}
			logger.Info("applied update",
				zap.String("C", registry.ConstructorName(update.Constructor)),
				zap.Int64("UpdateID", update.UpdateID),
			)
			switch update.Constructor {
			case msg.C_UpdateMessageID:
			default:
				udpContainer.Updates = append(udpContainer.Updates, externalHandlerUpdates...)
			}
		} else {
			udpContainer.Updates = append(udpContainer.Updates, update)
		}
		if update.UpdateID != 0 {
			_ = ctrl.SetUpdateID(update.UpdateID)
		}
	}

	udpContainer.Length = int32(len(udpContainer.Updates))
	uiexec.ExecUpdate(msg.C_UpdateContainer, updateContainer)
	return
}

// ResetIDs reset updateID
func (ctrl *Controller) ResetIDs() {
	_ = ctrl.SetUpdateID(0)
	ctrl.SetUserID(0)
}

// ContactsImport executes ContactsImport rpc commands
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
				logger.Error("got timeout on ContactsImport")
			},
			func(m *rony.MessageEnvelope) {
				defer wg.Done()
				switch m.Constructor {
				case msg.C_ContactsImported:
					x := &msg.ContactsImported{}
					err := x.Unmarshal(m.Message)
					if err != nil {
						logger.Error("got error on ContactsImport when unmarshal", zap.Error(err))
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
						logger.Warn("got error response from server, will retry",
							zap.String("Code", x.Code), zap.String("Item", x.Items),
						)
						time.Sleep(time.Second)
					}
					if maxTry--; maxTry < 0 {
						keepGoing = false
					}
				default:
					logger.Error("expected ContactsImported but we got something else!!!",
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
		uiexec.ExecDataSynced(false, true, false)
	}
	return
}

func (ctrl *Controller) GetSyncStatus() domain.SyncStatus {
	return ctrl.syncStatus
}
