package syncCtrl

import (
	"encoding/json"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
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
				time.Sleep(time.Duration(ronak.RandomInt(2000)) * time.Millisecond)
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

func (ctrl *Controller) AuthRecall(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
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
				time.Sleep(time.Duration(ronak.RandomInt(2000)) * time.Millisecond)
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
}

func (ctrl *Controller) GetServerTime(waitGroup *sync.WaitGroup) {
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

}

func (ctrl *Controller) GetAllDialogs(waitGroup *sync.WaitGroup, offset int32, limit int32) {
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
			ctrl.GetAllDialogs(waitGroup, offset, limit)
		},
		func(m *msg.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_Error:
				logs.Error("SyncCtrl got error response on MessagesGetDialogs",
					zap.String("Error", domain.ParseServerError(m.Message).Error()),
				)
				ctrl.GetAllDialogs(waitGroup, offset, limit)
			case msg.C_MessagesDialogs:
				x := new(msg.MessagesDialogs)
				err := x.Unmarshal(m.Message)
				if err != nil {
					logs.Error("SyncCtrl cannot unmarshal server response on MessagesGetDialogs", zap.Error(err))
					return
				}

				if x.Count > offset+limit {
					ctrl.GetAllDialogs(waitGroup, offset+limit, limit)
				} else {
					waitGroup.Done()
				}
			}
		},
		false,
	)
}

func (ctrl *Controller) GetContacts(waitGroup *sync.WaitGroup) {
	logs.Debug("SyncCtrl calls getContacts")
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
				errMsg := new(msg.Error)
				_ = errMsg.Unmarshal(m.Message)
				ctrl.GetContacts(waitGroup)
			default:
				waitGroup.Done()
			}
			// Controller applier will take care of this
		},
		false,
	)
}

func (ctrl *Controller) GetUpdateState(waitGroup *sync.WaitGroup) (updateID int64, err error) {
	logs.Debug("SyncCtrl calls getUpdateState")
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
	return
}

func (ctrl *Controller) SendGetUsers(waitGroup *sync.WaitGroup) {
	// TODO:: this is for not-stored users in the db
}