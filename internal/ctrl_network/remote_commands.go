package networkCtrl

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"github.com/ronaksoft/rony"
	"go.uber.org/zap"
)

/*
   Creation Time: 2020 - Dec - 15
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func (ctrl *Controller) GetDHGroups() (err error) {
	logs.Info("SyncCtrl call GetSystemDHGroups")
	req := &msg.SystemGetDHGroups{}
	reqBytes, _ := req.Marshal()
	ctrl.RealtimeCommand(
		&rony.MessageEnvelope{
			Constructor: msg.C_SystemGetDHGroups,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     reqBytes,
		},
		func() {

		},
		func(m *rony.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_SystemDHGroups:
				x := &msg.SystemDHGroups{}
				err = x.Unmarshal(m.Message)
				if err != nil {
					logs.Error("SyncCtrl couldn't unmarshal SystemGetServerTime response", zap.Error(err))
					return
				}

				logs.Debug("SyncCtrl received SystemServerTime",
					zap.Duration("Difference", domain.TimeDelta),
				)
			case rony.C_Error:
				logs.Warn("We received error on GetSystemServerTime", zap.Error(domain.ParseServerError(m.Message)))
				err = domain.ParseServerError(m.Message)
			}
		}, true, false,
	)
	return
}

func (ctrl *Controller) GetPublicKeys() (err error) {
	logs.Info("SyncCtrl call GetSystemPublicKeys")
	req := &msg.SystemGetPublicKeys{}
	reqBytes, _ := req.Marshal()
	ctrl.RealtimeCommand(
		&rony.MessageEnvelope{
			Constructor: msg.C_SystemGetPublicKeys,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     reqBytes,
		},
		func() {

		},
		func(m *rony.MessageEnvelope) {

		}, true, false,
	)
	return
}
