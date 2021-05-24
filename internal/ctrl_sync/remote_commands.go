package syncCtrl

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/uiexec"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/registry"
	"github.com/ronaksoft/rony/tools"
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
	logger.Info("call GetServerSalt")
	serverSaltReq := &msg.SystemGetSalts{}
	serverSaltReqBytes, _ := serverSaltReq.Marshal()

	ctrl.networkCtrl.WebsocketCommand(
		&rony.MessageEnvelope{
			Constructor: msg.C_SystemGetSalts,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     serverSaltReqBytes,
		},
		func() {
			time.Sleep(time.Duration(domain.RandomInt(2000)) * time.Millisecond)
		},
		func(m *rony.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_SystemSalts:
				logger.Debug("received SystemSalts")
			case rony.C_Error:
				e := new(rony.Error)
				_ = m.Unmarshal(m.Message)
				logger.Error("received error response for SystemGetSalts (Error)",
					zap.String("Code", e.Code),
					zap.String("Item", e.Items),
				)
				time.Sleep(time.Second)
			}
		},
		false,
		domain.RequestSkipFlusher,
	)
}

func (ctrl *Controller) GetSystemConfig() {
	logger.Info("call SystemGetConfig")
	req := &msg.SystemGetConfig{}
	reqBytes, _ := req.Marshal()

	ctrl.networkCtrl.WebsocketCommand(
		&rony.MessageEnvelope{
			Constructor: msg.C_SystemGetConfig,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     reqBytes,
		},
		func() {
			time.Sleep(time.Duration(domain.RandomInt(2000)) * time.Millisecond)
		},
		func(m *rony.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_SystemConfig:
				logger.Debug("received SystemConfig")
			case rony.C_Error:
				e := new(rony.Error)
				_ = m.Unmarshal(m.Message)
				logger.Error("received error response for SystemGetSalts (Error)",
					zap.String("Code", e.Code),
					zap.String("Item", e.Items),
				)
				time.Sleep(time.Second)
			}
		},
		false,
		domain.RequestSkipFlusher,
	)
}

func (ctrl *Controller) AuthRecall(caller string) (updateID int64, err error) {

	logger.Info("call AuthRecall", zap.String("Caller", caller))
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
	ctrl.networkCtrl.WebsocketCommand(
		&rony.MessageEnvelope{
			Constructor: msg.C_AuthRecall,
			RequestID:   reqID,
			Message:     reqBytes,
		},
		func() {
			logger.Warn("AuthRecall Timeout",
				zap.Uint64("ReqID", reqID),
				zap.Int64("AuthID", ctrl.connInfo.PickupAuthID()),
				zap.Int64("UserID", ctrl.connInfo.PickupUserID()),
			)
			err = domain.ErrRequestTimeout
			time.Sleep(time.Duration(domain.RandomInt(2000)) * time.Millisecond)
		},
		func(m *rony.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_AuthRecalled:
				x := &msg.AuthRecalled{}
				err = x.Unmarshal(m.Message)
				if err != nil {
					return
				}
				logger.Debug("received AuthRecalled")
				updateID = x.UpdateID

				// Update the time difference between client & server
				clientTime := time.Now().Unix()
				serverTime := x.Timestamp
				domain.TimeDelta = time.Duration(serverTime-clientTime) * time.Second

				ctrl.appUpdateCallback(x.CurrentVersion, x.Available, x.Force)
			case rony.C_Error:
				err = domain.ParseServerError(m.Message)
			default:
				logger.Error("did not received expected response for AuthRecall",
					zap.String("C", registry.ConstructorName(m.Constructor)),
				)
				err = domain.ErrInvalidConstructor
			}
		},
		false,
		domain.RequestSkipFlusher,
	)

	// Set the flag for network controller
	ctrl.networkCtrl.SetAuthRecalled(true)
	return
}

func (ctrl *Controller) GetServerTime() (err error) {
	logger.Info("call GetServerTime")
	timeReq := &msg.SystemGetServerTime{}
	timeReqBytes, _ := timeReq.Marshal()
	ctrl.networkCtrl.WebsocketCommand(
		&rony.MessageEnvelope{
			Constructor: msg.C_SystemGetServerTime,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     timeReqBytes,
		},
		func() {
			err = domain.ErrRequestTimeout
		},
		func(m *rony.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_SystemServerTime:
				x := new(msg.SystemServerTime)
				err = x.Unmarshal(m.Message)
				if err != nil {
					logger.Error("couldn't unmarshal SystemGetServerTime response", zap.Error(err))
					return
				}
				clientTime := time.Now().Unix()
				serverTime := x.Timestamp
				domain.TimeDelta = time.Duration(serverTime-clientTime) * time.Second

				logger.Debug("received SystemServerTime",
					zap.Int64("ServerTime", serverTime),
					zap.Int64("ClientTime", clientTime),
					zap.Duration("Difference", domain.TimeDelta),
				)
			case rony.C_Error:
				logger.Warn("received error on GetSystemServerTime", zap.Error(domain.ParseServerError(m.Message)))
				err = domain.ParseServerError(m.Message)
			}
		},
		false,
		domain.RequestSkipFlusher,
	)
	return
}

func (ctrl *Controller) GetAllDialogs(waitGroup *sync.WaitGroup, teamID int64, teamAccess uint64, offset int32, limit int32) {
	logger.Info("calls GetAllDialogs",
		zap.Int32("Offset", offset),
		zap.Int32("Limit", limit),
	)
	req := &msg.MessagesGetDialogs{
		Limit:  limit,
		Offset: offset,
	}
	reqBytes, _ := req.Marshal()
	ctrl.queueCtrl.EnqueueCommand(
		&rony.MessageEnvelope{
			Header:      domain.TeamHeader(teamID, teamAccess),
			Constructor: msg.C_MessagesGetDialogs,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     reqBytes,
		},
		func() {
			// If timeout, then retry the request
			logger.Warn("Timeout! on GetAllDialogs, retrying ...")
			_, _ = ctrl.AuthRecall("GetAllDialogs")
			ctrl.GetAllDialogs(waitGroup, teamID, teamAccess, offset, limit)
		},
		func(m *rony.MessageEnvelope) {
			switch m.Constructor {
			case rony.C_Error:
				logger.Error("got error response on MessagesGetDialogs", zap.Error(domain.ParseServerError(m.Message)))
				x := &rony.Error{}
				_ = x.Unmarshal(m.Message)
				if x.Code == msg.ErrCodeUnavailable && x.Items == msg.ErrItemUserID {
					waitGroup.Done()
					return
				} else if x.Code == msg.ErrCodeRateLimit {
					time.Sleep(time.Second * time.Duration(tools.StrToInt64(x.Items)))
				}
				ctrl.GetAllDialogs(waitGroup, teamID, teamAccess, offset, limit)

			case msg.C_MessagesDialogs:
				x := msg.MessagesDialogs{}
				err := x.Unmarshal(m.Message)
				if err != nil {
					logger.Error("cannot unmarshal server response on MessagesGetDialogs", zap.Error(err))
					return
				}

				if x.Count > offset+limit {
					ctrl.GetAllDialogs(waitGroup, teamID, teamAccess, offset+limit, limit)
				} else {
					waitGroup.Done()
					uiexec.ExecDataSynced(true, false, false)
				}
			}
		},
		false,
	)
}

func (ctrl *Controller) GetAllTopPeers(
	waitGroup *sync.WaitGroup, teamID int64, teamAccess uint64, cat msg.TopPeerCategory, offset int32, limit int32,
) {
	logger.Info("calls GetAllTopPeers",
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
		&rony.MessageEnvelope{
			Header:      domain.TeamHeader(teamID, teamAccess),
			Constructor: msg.C_ContactsGetTopPeers,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     reqBytes,
		},
		func() {
			// If timeout, then retry the request
			logger.Warn("Timeout! on GetAllTopPeers, retrying ...", zap.String("Cat", cat.String()))
			_, _ = ctrl.AuthRecall("GetAllTopPeers")
			ctrl.GetAllTopPeers(waitGroup, teamID, teamAccess, cat, offset, limit)
		},
		func(m *rony.MessageEnvelope) {
			switch m.Constructor {
			case rony.C_Error:
				logger.Error("got error response on ContactsGetTopPeers", zap.Error(domain.ParseServerError(m.Message)))
				x := &rony.Error{}
				_ = x.Unmarshal(m.Message)
				switch {
				case domain.CheckError(x, msg.ErrCodeUnavailable, msg.ErrItemUserID):
					fallthrough
				case domain.CheckError(x, msg.ErrCodeInvalid, msg.ErrItemAccessHash):
					fallthrough
				case domain.CheckErrorCode(x, msg.ErrCodeAccess):
					waitGroup.Done()
					return
				case domain.CheckErrorCode(x, msg.ErrCodeRateLimit):
					time.Sleep(time.Second * time.Duration(tools.StrToInt64(x.Items)))
				}
				ctrl.GetAllTopPeers(waitGroup, teamID, teamAccess, cat, offset, limit)
			case msg.C_ContactsTopPeers:
				x := msg.ContactsTopPeers{}
				err := x.Unmarshal(m.Message)
				if err != nil {
					logger.Error("cannot unmarshal server response on MessagesGetDialogs", zap.Error(err))
					return
				}

				if len(x.Peers) >= int(limit) {
					ctrl.GetAllTopPeers(waitGroup, teamID, teamAccess, cat, offset+limit, limit)
				} else {
					waitGroup.Done()
					uiexec.ExecDataSynced(true, false, false)
				}
			}
		},
		false,
	)
}

func (ctrl *Controller) GetLabels(waitGroup *sync.WaitGroup, teamID int64, teamAccess uint64) {
	logger.Info("calls GetLabels")
	req := &msg.LabelsGet{}
	reqBytes, _ := req.Marshal()
	ctrl.queueCtrl.EnqueueCommand(
		&rony.MessageEnvelope{
			Header:      domain.TeamHeader(teamID, teamAccess),
			Constructor: msg.C_LabelsGet,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     reqBytes,
		},
		func() {
			// If timeout, then retry the request
			logger.Warn("Timeout! on LabelsGet, retrying ...")
			_, _ = ctrl.AuthRecall("LabelsGet")
			ctrl.GetLabels(waitGroup, teamID, teamAccess)
		},
		func(m *rony.MessageEnvelope) {
			switch m.Constructor {
			case rony.C_Error:
				logger.Error("got error response on LabelsGet", zap.Error(domain.ParseServerError(m.Message)))
				x := &rony.Error{}
				_ = x.Unmarshal(m.Message)
				switch {
				case domain.CheckError(x, msg.ErrCodeUnavailable, msg.ErrItemUserID):
					fallthrough
				case domain.CheckError(x, msg.ErrCodeInvalid, msg.ErrItemAccessHash):
					fallthrough
				case domain.CheckErrorCode(x, msg.ErrCodeAccess):
					waitGroup.Done()
					return
				case domain.CheckErrorCode(x, msg.ErrCodeRateLimit):
					time.Sleep(time.Second * time.Duration(tools.StrToInt64(x.Items)))
				}
				ctrl.GetLabels(waitGroup, teamID, teamAccess)
			case msg.C_LabelsMany:
				waitGroup.Done()
			}
		},
		false,
	)
}

func (ctrl *Controller) GetContacts(waitGroup *sync.WaitGroup, teamID int64, teamAccess uint64) {
	logger.Debug("calls GetContacts")

	contactsGetHash, _ := repo.System.LoadInt(domain.GetContactsGetHashKey(teamID))
	req := &msg.ContactsGet{
		Crc32Hash: uint32(contactsGetHash),
	}
	reqBytes, _ := req.Marshal()
	ctrl.queueCtrl.EnqueueCommand(
		&rony.MessageEnvelope{
			Header:      domain.TeamHeader(teamID, teamAccess),
			Constructor: msg.C_ContactsGet,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     reqBytes,
		},
		func() {
			ctrl.GetContacts(waitGroup, teamID, teamAccess)
		},
		func(m *rony.MessageEnvelope) {
			switch m.Constructor {
			case rony.C_Error:
				x := new(rony.Error)
				_ = x.Unmarshal(m.Message)
				if x.Code == msg.ErrCodeUnavailable && x.Items == msg.ErrItemUserID {
					waitGroup.Done()
					return
				} else if x.Code == msg.ErrCodeRateLimit {
					time.Sleep(time.Second * time.Duration(tools.StrToInt64(x.Items)))
				}
				ctrl.GetContacts(waitGroup, teamID, teamAccess)

			default:
				waitGroup.Done()
			}
			// Controller applier will take care of this
		},
		false,
	)
}

func (ctrl *Controller) Logout(waitGroup *sync.WaitGroup, retry int) {
	if retry <= 0 {
		waitGroup.Done()
		return
	}
	requestID := domain.SequentialUniqueID()
	req := &msg.AuthLogout{}
	reqBytes, _ := req.Marshal()
	go ctrl.networkCtrl.WebsocketCommand(
		&rony.MessageEnvelope{
			Constructor: msg.C_AuthLogout,
			RequestID:   uint64(requestID),
			Message:     reqBytes,
		},
		func() {
			logger.Info("Logout Request was timeout, will retry")
			ctrl.networkCtrl.Reconnect()
			ctrl.Logout(waitGroup, retry-1)
		},
		func(m *rony.MessageEnvelope) {
			switch m.Constructor {
			case rony.C_Error:
				x := &rony.Error{}
				_ = x.Unmarshal(m.Message)
				logger.Warn("got error on AuthLogout", zap.String("Code", x.Code), zap.String("Item", x.Items))
				ctrl.Logout(waitGroup, retry-1)
			default:
				waitGroup.Done()
			}
			// Controller applier will take care of this
		},
		false, domain.RequestSkipFlusher,
	)
}

func (ctrl *Controller) UpdateStatus(online bool) {
	req := &msg.AccountUpdateStatus{
		Online: online,
	}
	reqBytes, _ := req.Marshal()
	go ctrl.networkCtrl.WebsocketCommand(
		&rony.MessageEnvelope{
			Constructor: msg.C_AccountUpdateStatus,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     reqBytes,
		},
		func() {
			return
		},
		func(m *rony.MessageEnvelope) {
			switch m.Constructor {
			case rony.C_Error:
				x := &rony.Error{}
				_ = x.Unmarshal(m.Message)
				if !(x.Code == msg.ErrCodeUnavailable && x.Items == msg.ErrItemUserID) {
					time.Sleep(time.Second)
					ctrl.UpdateStatus(online)
				}
			}
			// Controller applier will take care of this
		},
		false,
		0,
	)
}

func (ctrl *Controller) UploadUsage() error {
	logger.Debug("calls SystemUploadUsage")
	req := &msg.SystemUploadUsage{}
	req.Usage = append(req.Usage)
	reqBytes, _ := req.Marshal()
	ctrl.queueCtrl.EnqueueCommand(
		&rony.MessageEnvelope{
			Constructor: msg.C_SystemUploadUsage,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     reqBytes,
		},
		func() {},
		func(m *rony.MessageEnvelope) {},
		false,
	)
	return nil
}
