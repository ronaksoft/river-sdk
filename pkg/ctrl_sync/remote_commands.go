package syncCtrl

import (
	msg "git.ronaksoftware.com/ronak/riversdk/msg/chat"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	ronak "git.ronaksoftware.com/ronak/toolbox"
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
	serverSaltReq := new(msg.SystemGetSalts)
	serverSaltReqBytes, _ := serverSaltReq.Marshal()

	ctrl.queueCtrl.RealtimeCommand(
		uint64(domain.SequentialUniqueID()),
		msg.C_SystemGetSalts,
		serverSaltReqBytes,
		func() {
			time.Sleep(time.Duration(ronak.RandomInt(2000)) * time.Millisecond)
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
		reqID,
		msg.C_AuthRecall,
		reqBytes,
		func() {
			logs.Warn("AuthRecall Timeout",
				zap.Uint64("ReqID", reqID),
				zap.Int64("AuthID", ctrl.connInfo.PickupAuthID()),
				zap.Int64("UserID", ctrl.connInfo.PickupUserID()),
			)
			err = domain.ErrRequestTimeout
			time.Sleep(time.Duration(ronak.RandomInt(2000)) * time.Millisecond)
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
					zap.String("Constructor", msg.ConstructorNames[m.Constructor]),
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
	timeReq := new(msg.SystemGetServerTime)
	timeReqBytes, _ := timeReq.Marshal()
	ctrl.queueCtrl.RealtimeCommand(
		uint64(domain.SequentialUniqueID()),
		msg.C_SystemGetServerTime,
		timeReqBytes,
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

func (ctrl *Controller) GetAllDialogs(waitGroup *sync.WaitGroup, offset int32, limit int32) {
	logs.Info("SyncCtrl calls GetAllDialogs",
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
			_, _ = ctrl.AuthRecall("GetAllDialogs")
			ctrl.GetAllDialogs(waitGroup, offset, limit)
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
					ctrl.GetAllDialogs(waitGroup, offset, limit)
				}
			case msg.C_MessagesDialogs:
				x := msg.MessagesDialogs{}
				err := x.Unmarshal(m.Message)
				if err != nil {
					logs.Error("SyncCtrl cannot unmarshal server response on MessagesGetDialogs", zap.Error(err))
					return
				}

				if x.Count > offset+limit {
					ctrl.GetAllDialogs(waitGroup, offset+limit, limit)
				} else {
					waitGroup.Done()
					updateUI(ctrl, true, false)
				}
			}
		},
		false,
	)
}

func (ctrl *Controller) GetContacts(waitGroup *sync.WaitGroup) {
	logs.Debug("SyncCtrl calls GetContacts")
	req := new(msg.ContactsGet)
	reqBytes, _ := req.Marshal()
	ctrl.queueCtrl.EnqueueCommand(
		uint64(domain.SequentialUniqueID()),
		msg.C_ContactsGet,
		reqBytes,
		func() {
			ctrl.GetContacts(waitGroup)
		},
		func(m *msg.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_Error:
				x := new(msg.Error)
				_ = x.Unmarshal(m.Message)
				if x.Code == msg.ErrCodeUnavailable && x.Items == msg.ErrItemUserID {
					waitGroup.Done()
				} else {
					ctrl.GetContacts(waitGroup)
				}
			default:
				waitGroup.Done()
			}
			// Controller applier will take care of this
		},
		false,
	)
}

func (ctrl *Controller) GetUpdateState() (updateID int64, err error) {
	logs.Debug("SyncCtrl calls GetUpdateState")
	updateID = 0
	if !ctrl.networkCtrl.Connected() {
		return -1, domain.ErrNoConnection
	}

	req := new(msg.UpdateGetState)
	reqBytes, _ := req.Marshal()

	ctrl.queueCtrl.RealtimeCommand(
		uint64(domain.SequentialUniqueID()),
		msg.C_UpdateGetState,
		reqBytes,
		func() {
			err = domain.ErrRequestTimeout
		},
		func(m *msg.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_UpdateState:
				x := new(msg.UpdateState)
				_ = x.Unmarshal(m.Message)
				updateID = x.UpdateID
				logs.Debug("SyncCtrl received UpdateState", zap.Int64("UpdateID", updateID))
			case msg.C_Error:
				err = domain.ParseServerError(m.Message)
			}
		},
		true,
		false,
	)
	return
}
