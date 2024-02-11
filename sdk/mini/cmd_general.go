package mini

import (
    "runtime"
    "sync"
    "time"

    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/river-sdk/internal/logs"
    "github.com/ronaksoft/river-sdk/internal/minirepo"
    "github.com/ronaksoft/river-sdk/internal/request"
    "github.com/ronaksoft/rony/registry"
    "go.uber.org/zap"
)

/*
   Creation Time: 2021 - Apr - 30
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

// AppKill must be called when app is closed
func (r *River) AppKill() {
    r.network.Stop()
}

// AppStart must be called when app is started
func (r *River) AppStart() error {
    runtime.GOMAXPROCS(runtime.NumCPU() * 2)

    logs.SetSentry(r.ConnInfo.AuthID, r.ConnInfo.UserID, r.sentryDSN)
    logger.Info("MiniRiver Starting")

    minirepo.MustInit(r.dbPath)

    // Update Authorizations
    r.network.SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey)

    // Start Controllers
    r.network.Start()

    domain.StartTime = time.Now()
    domain.WindowLog = func(txt string) {}
    logger.Info("MiniRiver Started")

    err := r.syncServerTime()
    if err != nil {
        logger.Warn("MiniRiver got error on get server time", zap.Error(err))
    }

    // Run update/message processors in background
    go r.messageReceiver()
    go r.updateReceiver()

    if r.getLastUpdateID(0) == 0 {
        // run in sync for the first time
        wg := sync.WaitGroup{}
        wg.Add(3)
        go func() {
            r.syncTeams()
            wg.Done()
        }()
        go func() {
            r.syncContacts(0, 0)
            wg.Done()
        }()
        go func() {
            r.syncDialogs(0, 0)
            wg.Done()
        }()
        wg.Wait()
    } else {
        // run in background
        go r.syncContacts(0, 0)
        go r.syncDialogs(0, 0)
        go r.syncTeams()
    }

    return nil
}

// ExecuteCommand is a wrapper function to pass the request to the queueController, to be passed to networkController for final
// delivery to the server. SDK uses GetCurrentTeam() to detect the targeted team of the request
func (r *River) ExecuteCommand(constructor int64, commandBytes []byte, delegate RequestDelegate) (requestID int64, err error) {
    requestID = domain.SequentialUniqueID()
    err = r.executeCommand(
        request.DelegateAdapter(
            domain.GetCurrTeamID(), domain.GetCurrTeamAccess(), uint64(requestID), constructor, commandBytes, delegate, delegate.OnProgress,
        ),
    )
    return requestID, err
}

// ExecuteCommandWithTeam is similar to ExecuteTeam but explicitly defines the target team
func (r *River) ExecuteCommandWithTeam(
        teamID, accessHash, constructor int64, commandBytes []byte, delegate RequestDelegate,
) (requestID int64, err error) {
    requestID = domain.SequentialUniqueID()
    err = r.executeCommand(
        request.DelegateAdapter(
            teamID, uint64(accessHash), uint64(requestID), constructor, commandBytes, delegate, delegate.OnProgress,
        ),
    )
    return requestID, err
}

func (r *River) executeCommand(reqCB request.Callback) (err error) {
    if registry.ConstructorName(reqCB.Constructor()) == "" {
        return domain.ErrInvalidConstructor
    }

    logger.Debug("executes command",
        zap.Uint64("ReqID", reqCB.RequestID()),
        zap.String("C", registry.ConstructorName(reqCB.Constructor())),
        zap.String("Flags", request.DelegateFlagToString(reqCB.Flags())),
    )

    serverForce := reqCB.Flags()&request.ServerForced != 0

    // If the constructor is a local command then
    handler, ok := r.localCommands[reqCB.Constructor()]
    if ok && !serverForce {
        r.executeLocalCommand(handler, reqCB)
        return
    }

    // If we reached here, then execute the remote commands
    r.executeRemoteCommand(reqCB)

    return
}
func (r *River) executeLocalCommand(handler request.LocalHandler, reqCB request.Callback) {
    logger.Info("execute local command",
        zap.Uint64("ReqID", reqCB.RequestID()),
        zap.String("C", registry.ConstructorName(reqCB.Constructor())),
        zap.String("Flags", request.DelegateFlagToString(reqCB.Flags())),
    )

    handler(reqCB)
}
func (r *River) executeRemoteCommand(reqCB request.Callback) {
    logger.Info("execute remote command",
        zap.Uint64("ReqID", reqCB.RequestID()),
        zap.String("C", registry.ConstructorName(reqCB.Constructor())),
        zap.String("Flags", request.DelegateFlagToString(reqCB.Flags())),
    )

    r.network.HttpCommand(nil, reqCB)
}

func (r *River) SetTeam(teamID int64, teamAccessHash int64) {
    domain.SetCurrentTeam(teamID, uint64(teamAccessHash))
    if r.getLastUpdateID(teamID) == 0 {
        // run in sync for the first time
        wg := sync.WaitGroup{}
        wg.Add(2)
        go func() {
            r.syncContacts(teamID, uint64(teamAccessHash))
            wg.Done()
        }()
        go func() {
            r.syncDialogs(teamID, uint64(teamAccessHash))
            wg.Done()
        }()
        wg.Wait()
    } else {
        // run in background
        go r.syncContacts(teamID, uint64(teamAccessHash))
        go r.syncDialogs(teamID, uint64(teamAccessHash))
    }

}
