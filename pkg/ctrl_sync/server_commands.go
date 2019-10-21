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

func (ctrl *Controller) SendGetUsers(waitGroup *sync.WaitGroup) {
	// TODO:: this is for not-stored users in the db
}