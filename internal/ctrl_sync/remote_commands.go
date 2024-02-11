package syncCtrl

import (
    "sync"
    "time"

    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/river-sdk/internal/repo"
    "github.com/ronaksoft/river-sdk/internal/request"
    "github.com/ronaksoft/river-sdk/internal/uiexec"
    "github.com/ronaksoft/rony"
    "github.com/ronaksoft/rony/registry"
    "github.com/ronaksoft/rony/tools"
    "go.uber.org/zap"
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
    ctrl.networkCtrl.WebsocketCommand(
        request.NewCallback(
            0, 0, domain.NextRequestID(), msg.C_SystemGetSalts, &msg.SystemGetSalts{},
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
            }, nil,
            false, request.SkipFlusher, 0,
        ),
    )
}

func (ctrl *Controller) GetSystemConfig() {
    logger.Info("call SystemGetConfig")
    ctrl.networkCtrl.WebsocketCommand(
        request.NewCallback(
            0, 0, domain.NextRequestID(), msg.C_SystemGetConfig, &msg.SystemGetConfig{},
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
            }, nil,
            false, 0, 0,
        ),
    )
}

func (ctrl *Controller) AuthRecall(caller string) (updateID int64, err error) {
    logger.Info("call AuthRecall", zap.String("Caller", caller))
    req := &msg.AuthRecall{
        ClientID:   0,
        Version:    0,
        AppVersion: domain.ClientVersion,
        Platform:   domain.ClientPlatform,
        Vendor:     domain.ClientVendor,
        OSVersion:  domain.ClientOS,
    }
    ctrl.networkCtrl.WebsocketCommand(
        request.NewCallback(
            0, 0, domain.NextRequestID(), msg.C_AuthRecall, req,
            func() {
                logger.Warn("AuthRecall Timeout",
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
                    clientTime := tools.TimeUnix()
                    serverTime := x.Timestamp
                    domain.TimeDelta = time.Duration(serverTime-clientTime) * time.Second

                    // Set the flag for network controller
                    ctrl.networkCtrl.SetAuthRecalled(true)

                    ctrl.appUpdateCallback(x.CurrentVersion, x.Available, x.Force)
                case rony.C_Error:
                    err = domain.ParseServerError(m.Message)
                default:
                    logger.Error("did not received expected response for AuthRecall",
                        zap.String("C", registry.ConstructorName(m.Constructor)),
                    )
                    err = domain.ErrInvalidConstructor
                }
            }, nil,
            false, request.SkipFlusher, 0,
        ),
    )

    return
}

func (ctrl *Controller) GetServerTime() (err error) {
    logger.Info("calls GetServerTime")
    ctrl.networkCtrl.WebsocketCommand(
        request.NewCallback(
            0, 0, domain.NextRequestID(), msg.C_SystemGetServerTime, &msg.SystemGetServerTime{},
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
            }, nil,
            false, request.SkipFlusher, 0,
        ),
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

    ctrl.networkCtrl.WebsocketCommand(
        request.NewCallback(
            teamID, teamAccess, domain.NextRequestID(), msg.C_MessagesGetDialogs, req,
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
            }, nil, false, 0, 0,
        ),
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
    ctrl.networkCtrl.WebsocketCommand(
        request.NewCallback(
            teamID, teamAccess, domain.NextRequestID(), msg.C_ContactsGetTopPeers, req,
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
            }, nil,
            false, 0, 0,
        ),
    )
}

func (ctrl *Controller) GetLabels(waitGroup *sync.WaitGroup, teamID int64, teamAccess uint64) {
    logger.Info("calls GetLabels")
    ctrl.networkCtrl.WebsocketCommand(
        request.NewCallback(
            teamID, teamAccess, domain.NextRequestID(), msg.C_LabelsGet, &msg.LabelsGet{},
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
            }, nil,
            false, 0, 0,
        ),
    )
}

func (ctrl *Controller) GetContacts(waitGroup *sync.WaitGroup, teamID int64, teamAccess uint64) {
    logger.Debug("calls GetContacts")

    contactsGetHash, _ := repo.System.LoadInt(domain.GetContactsGetHashKey(teamID))
    req := &msg.ContactsGet{
        Crc32Hash: uint32(contactsGetHash),
    }
    ctrl.networkCtrl.WebsocketCommand(
        request.NewCallback(
            teamID, teamAccess, domain.NextRequestID(), msg.C_ContactsGet, req,
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
            }, nil,
            false, 0, 0,
        ),
    )
}

func (ctrl *Controller) Logout(waitGroup *sync.WaitGroup, retry int) {
    if retry <= 0 {
        waitGroup.Done()
        return
    }
    go ctrl.networkCtrl.WebsocketCommand(
        request.NewCallback(
            0, 0, domain.NextRequestID(), msg.C_AuthLogout, &msg.AuthLogout{},
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
            }, nil,
            false, request.SkipFlusher, 0,
        ),
    )
}

func (ctrl *Controller) UpdateStatus(online bool) {
    req := &msg.AccountUpdateStatus{
        Online: online,
    }
    retry := 3
    go ctrl.networkCtrl.WebsocketCommand(
        request.NewCallback(
            0, 0, domain.NextRequestID(), msg.C_AccountUpdateStatus, req,
            func() {
                if retry--; retry > 0 {
                    time.Sleep(time.Second)
                    ctrl.UpdateStatus(online)
                }
            },
            func(m *rony.MessageEnvelope) {
                switch m.Constructor {
                case rony.C_Error:
                    x := &rony.Error{}
                    _ = x.Unmarshal(m.Message)
                    logger.Warn("got error on AccountUpdateStatus", zap.Error(x))
                }
                // Controller applier will take care of this
            }, nil,
            false, 0, 0,
        ),
    )
}
