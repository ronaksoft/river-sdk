package syncCtrl

import (
	"git.ronaksoft.com/river/msg/msg"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/pkg/domain"
	"git.ronaksoft.com/river/sdk/pkg/repo"
	"go.uber.org/zap"
	"sync"
	"time"
)

/*
   Creation Time: 2019 - Oct - 21
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func (ctrl *Controller) GetServerSalt() {
	logs.Info("SyncCtrl call GetServerSalt")
	serverSaltReq := &msg.SystemGetSalts{}
	serverSaltReqBytes, _ := serverSaltReq.Marshal()

	ctrl.queueCtrl.RealtimeCommand(
		&msg.MessageEnvelope{
			Constructor: msg.C_SystemGetSalts,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     serverSaltReqBytes,
		},
		func() {
			time.Sleep(time.Duration(domain.RandomInt(2000)) * time.Millisecond)
		},
		func(m *msg.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_SystemSalts:
				logs.Debug("SyncCtrl received SystemSalts")
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

func (ctrl *Controller) GetSystemConfig() {
	logs.Info("SyncCtrl call SystemGetConfig")
	req := &msg.SystemGetConfig{}
	reqBytes, _ := req.Marshal()

	ctrl.queueCtrl.RealtimeCommand(
		&msg.MessageEnvelope{
			Constructor: msg.C_SystemGetConfig,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     reqBytes,
		},
		func() {
			time.Sleep(time.Duration(domain.RandomInt(2000)) * time.Millisecond)
		},
		func(m *msg.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_SystemConfig:
				logs.Debug("SyncCtrl received SystemConfig")
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

func (ctrl *Controller) AuthRecall(caller string) (updateID int64, err error) {
	logs.Info("SyncCtrl call AuthRecall", zap.String("Caller", caller))
	req := msg.AuthRecall{
		ClientID:   0,
		Version:    0,
		AppVersion: domain.ClientVersion,
		Platform:   domain.ClientPlatform,
		Vendor:     domain.ClientVendor,
		OSVersion:  domain.ClientOS,
	}
	reqBytes, _ := req.Marshal()

	// this is priority command that should not passed to queue
	// after auth recall answer got back the queue should send its requests in order to get related updates
	reqID := uint64(domain.SequentialUniqueID())
	ctrl.queueCtrl.RealtimeCommand(
		&msg.MessageEnvelope{
			Constructor: msg.C_AuthRecall,
			RequestID:   reqID,
			Message:     reqBytes,
		},
		func() {
			logs.Warn("AuthRecall Timeout",
				zap.Uint64("ReqID", reqID),
				zap.Int64("AuthID", ctrl.connInfo.PickupAuthID()),
				zap.Int64("UserID", ctrl.connInfo.PickupUserID()),
			)
			err = domain.ErrRequestTimeout
			time.Sleep(time.Duration(domain.RandomInt(2000)) * time.Millisecond)
		},
		func(m *msg.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_AuthRecalled:
				x := &msg.AuthRecalled{}
				err = x.Unmarshal(m.Message)
				if err != nil {
					return
				}
				logs.Debug("SyncCtrl received AuthRecalled")
				updateID = x.UpdateID

				// Update the time difference between client & server
				clientTime := time.Now().Unix()
				serverTime := x.Timestamp
				domain.TimeDelta = time.Duration(serverTime-clientTime) * time.Second

				ctrl.appUpdateCallback(x.CurrentVersion, x.Available, x.Force)
			case msg.C_Error:
				err = domain.ParseServerError(m.Message)
			default:
				logs.Error("SyncCtrl did not received expected response for AuthRecall",
					zap.String("C", msg.ConstructorNames[m.Constructor]),
				)
				err = domain.ErrInvalidConstructor
			}
		},
		true,
		false,
	)
	return
}

func (ctrl *Controller) GetServerTime() (err error) {
	logs.Info("SyncCtrl call GetServerTime")
	timeReq := &msg.SystemGetServerTime{}
	timeReqBytes, _ := timeReq.Marshal()
	ctrl.queueCtrl.RealtimeCommand(
		&msg.MessageEnvelope{
			Constructor: msg.C_SystemGetServerTime,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     timeReqBytes,
		},
		func() {
			err = domain.ErrRequestTimeout
		},
		func(m *msg.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_SystemServerTime:
				x := new(msg.SystemServerTime)
				err = x.Unmarshal(m.Message)
				if err != nil {
					logs.Error("SyncCtrl couldn't unmarshal SystemGetServerTime response", zap.Error(err))
					return
				}
				clientTime := time.Now().Unix()
				serverTime := x.Timestamp
				domain.TimeDelta = time.Duration(serverTime-clientTime) * time.Second

				logs.Debug("SyncCtrl received SystemServerTime",
					zap.Int64("ServerTime", serverTime),
					zap.Int64("ClientTime", clientTime),
					zap.Duration("Difference", domain.TimeDelta),
				)

			case msg.C_Error:
				logs.Warn("We received error on GetSystemServerTime", zap.Error(domain.ParseServerError(m.Message)))
				err = domain.ParseServerError(m.Message)
			}
		},
		true, false,
	)
	return
}

func (ctrl *Controller) GetAllDialogs(waitGroup *sync.WaitGroup, team *msg.InputTeam, offset int32, limit int32) {
	logs.Info("SyncCtrl calls GetAllDialogs",
		zap.Int32("Offset", offset),
		zap.Int32("Limit", limit),
	)
	req := &msg.MessagesGetDialogs{
		Limit:  limit,
		Offset: offset,
	}
	reqBytes, _ := req.Marshal()
	ctrl.queueCtrl.EnqueueCommand(
		&msg.MessageEnvelope{
			Team:        team,
			Constructor: msg.C_MessagesGetDialogs,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     reqBytes,
		},
		func() {
			// If timeout, then retry the request
			logs.Warn("Timeout! on GetAllDialogs, retrying ...")
			_, _ = ctrl.AuthRecall("GetAllDialogs")
			ctrl.GetAllDialogs(waitGroup, team, offset, limit)
		},
		func(m *msg.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_Error:
				logs.Error("SyncCtrl got error response on MessagesGetDialogs", zap.Error(domain.ParseServerError(m.Message)))
				x := msg.Error{}
				_ = x.Unmarshal(m.Message)
				if x.Code == msg.ErrCodeUnavailable && x.Items == msg.ErrItemUserID {
					waitGroup.Done()
				} else {
					ctrl.GetAllDialogs(waitGroup, team, offset, limit)
				}
			case msg.C_MessagesDialogs:
				x := msg.MessagesDialogs{}
				err := x.Unmarshal(m.Message)
				if err != nil {
					logs.Error("SyncCtrl cannot unmarshal server response on MessagesGetDialogs", zap.Error(err))
					return
				}

				if x.Count > offset+limit {
					ctrl.GetAllDialogs(waitGroup, team, offset+limit, limit)
				} else {
					waitGroup.Done()
					forceUpdateUI(ctrl, true, false, false)
				}
			}
		},
		false,
	)
}

func (ctrl *Controller) GetAllTopPeers(waitGroup *sync.WaitGroup, team *msg.InputTeam, cat msg.TopPeerCategory, offset int32, limit int32) {
	logs.Info("SyncCtrl calls GetAllTopPeers",
		zap.Int32("Offset", offset),
		zap.Int32("Limit", limit),
	)
	req := &msg.ContactsGetTopPeers{
		Limit:    limit,
		Offset:   offset,
		Category: cat,
	}
	reqBytes, _ := req.Marshal()
	ctrl.queueCtrl.EnqueueCommand(
		&msg.MessageEnvelope{
			Team:        team,
			Constructor: msg.C_ContactsGetTopPeers,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     reqBytes,
		},
		func() {
			// If timeout, then retry the request
			logs.Warn("Timeout! on GetAllTopPeers, retrying ...", zap.String("Cat", cat.String()))
			_, _ = ctrl.AuthRecall("GetAllTopPeers")
			ctrl.GetAllTopPeers(waitGroup, team, cat, offset, limit)
		},
		func(m *msg.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_Error:
				logs.Error("SyncCtrl got error response on ContactsGetTopPeers", zap.Error(domain.ParseServerError(m.Message)))
				x := msg.Error{}
				_ = x.Unmarshal(m.Message)
				if x.Code == msg.ErrCodeUnavailable && x.Items == msg.ErrItemUserID {
					waitGroup.Done()
				} else {
					ctrl.GetAllTopPeers(waitGroup, team, cat, offset, limit)
				}
			case msg.C_ContactsTopPeers:
				x := msg.ContactsTopPeers{}
				err := x.Unmarshal(m.Message)
				if err != nil {
					logs.Error("SyncCtrl cannot unmarshal server response on MessagesGetDialogs", zap.Error(err))
					return
				}

				if len(x.Peers) >= int(limit) {
					ctrl.GetAllTopPeers(waitGroup, team, cat, offset+limit, limit)
				} else {
					waitGroup.Done()
					forceUpdateUI(ctrl, true, false, false)
				}
			}
		},
		false,
	)
}

func (ctrl *Controller) GetLabels(waitGroup *sync.WaitGroup, team *msg.InputTeam) {
	logs.Info("SyncCtrl calls GetLabels")
	req := &msg.LabelsGet{}
	reqBytes, _ := req.Marshal()
	ctrl.queueCtrl.EnqueueCommand(
		&msg.MessageEnvelope{
			Team:        team,
			Constructor: msg.C_LabelsGet,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     reqBytes,
		},
		func() {
			// If timeout, then retry the request
			logs.Warn("Timeout! on LabelsGet, retrying ...")
			_, _ = ctrl.AuthRecall("LabelsGet")
			ctrl.GetLabels(waitGroup, team)
		},
		func(m *msg.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_Error:
				logs.Error("SyncCtrl got error response on LabelsGet", zap.Error(domain.ParseServerError(m.Message)))
				x := msg.Error{}
				_ = x.Unmarshal(m.Message)
				if x.Code == msg.ErrCodeUnavailable && x.Items == msg.ErrItemUserID {
					waitGroup.Done()
				} else {
					ctrl.GetLabels(waitGroup, team)
				}
			case msg.C_LabelsMany:
				waitGroup.Done()
			}
		},
		false,
	)
}

func (ctrl *Controller) GetContacts(waitGroup *sync.WaitGroup, team *msg.InputTeam) {
	logs.Debug("SyncCtrl calls GetContacts")
	var teamID int64
	if team == nil {
		teamID = 0
	} else {
		teamID = team.ID
	}
	contactsGetHash, _ := repo.System.LoadInt(domain.GetContactsGetHashKey(teamID))
	req := &msg.ContactsGet{
		Crc32Hash: uint32(contactsGetHash),
	}
	reqBytes, _ := req.Marshal()
	ctrl.queueCtrl.EnqueueCommand(
		&msg.MessageEnvelope{
			Team:        team,
			Constructor: msg.C_ContactsGet,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     reqBytes,
		},
		func() {
			ctrl.GetContacts(waitGroup, team)
		},
		func(m *msg.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_Error:
				x := new(msg.Error)
				_ = x.Unmarshal(m.Message)
				if x.Code == msg.ErrCodeUnavailable && x.Items == msg.ErrItemUserID {
					waitGroup.Done()
				} else {
					ctrl.GetContacts(waitGroup, team)
				}
			default:
				waitGroup.Done()
			}
			// Controller applier will take care of this
		},
		false,
	)
}

func (ctrl *Controller) UpdateStatus(online bool) {
	req := &msg.AccountUpdateStatus{
		Online: online,
	}
	reqBytes, _ := req.Marshal()
	ctrl.queueCtrl.RealtimeCommand(
		&msg.MessageEnvelope{
			Team:        nil,
			Constructor: msg.C_AccountUpdateStatus,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     reqBytes,
		},
		func() {
			ctrl.UpdateStatus(online)
		},
		func(m *msg.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_Error:
				x := new(msg.Error)
				_ = x.Unmarshal(m.Message)
				if x.Code == msg.ErrCodeUnavailable && x.Items == msg.ErrItemUserID {
					return
				} else {
					ctrl.UpdateStatus(online)
				}
			default:
				return
			}
			// Controller applier will take care of this
		},
		false,
		false,
	)
}
func (ctrl *Controller) UploadUsage() error {
	logs.Debug("SyncCtrl calls SystemUploadUsage")
	req := &msg.SystemUploadUsage{}
	req.Usage = append(req.Usage)
	reqBytes, _ := req.Marshal()
	ctrl.queueCtrl.EnqueueCommand(
		&msg.MessageEnvelope{
			Constructor: msg.C_SystemUploadUsage,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     reqBytes,
		},
		func() {},
		func(m *msg.MessageEnvelope) {},
		false,
	)
	return nil
}
