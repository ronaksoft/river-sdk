package mini

import (
	"context"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	mon "git.ronaksoft.com/river/sdk/internal/monitoring"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/salt"
	riversdk "git.ronaksoft.com/river/sdk/sdk/prime"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/registry"
	"go.uber.org/zap"
	"runtime"
	"time"
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
func (r *River) AppKill() {}

// AppStart must be called when app is started
func (r *River) AppStart() error {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	logs.Info("Mini River Starting")
	logs.SetSentry(r.ConnInfo.AuthID, r.ConnInfo.UserID, r.sentryDSN)

	// Initialize DB replaced with ORM
	err := repo.InitRepo(r.dbPath, true)
	if err != nil {
		return err
	}

	repo.SetSelfUserID(r.ConnInfo.UserID)

	confBytes, _ := repo.System.LoadBytes("SysConfig")
	if confBytes != nil {
		domain.SysConfig.Reactions = domain.SysConfig.Reactions[:0]
		err := domain.SysConfig.Unmarshal(confBytes)
		if err != nil {
			logs.Warn("We could not unmarshal SysConfig", zap.Error(err))
		}
	}

	// Load the usage stats
	mon.LoadUsage()

	// Update Authorizations
	r.networkCtrl.SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey[:])

	// Update the current salt
	salt.UpdateSalt()

	// Start Controllers
	r.networkCtrl.Start()
	r.fileCtrl.Start()

	domain.StartTime = time.Now()
	domain.WindowLog = func(txt string) {}
	logs.Info("River Started")

	return nil
}

// ExecuteCommand is a wrapper function to pass the request to the queueController, to be passed to networkController for final
// delivery to the server. SDK uses GetCurrentTeam() to detect the targeted team of the request
func (r *River) ExecuteCommand(constructor int64, commandBytes []byte, delegate riversdk.RequestDelegate) (requestID int64, err error) {
	return r.executeCommand(domain.GetCurrTeamID(), domain.GetCurrTeamAccess(), constructor, commandBytes, delegate)
}

// ExecuteCommandWithTeam is similar to ExecuteTeam but explicitly defines the target team
func (r *River) ExecuteCommandWithTeam(teamID, accessHash, constructor int64, commandBytes []byte, delegate riversdk.RequestDelegate) (requestID int64, err error) {
	return r.executeCommand(teamID, uint64(accessHash), constructor, commandBytes, delegate)
}

func (r *River) executeCommand(
	teamID int64, teamAccess uint64, constructor int64, commandBytes []byte, delegate riversdk.RequestDelegate,
) (requestID int64, err error) {
	if registry.ConstructorName(constructor) == "" {
		return 0, domain.ErrInvalidConstructor
	}

	// Timeout Callback
	timeoutCallback := func() {
		err = domain.ErrRequestTimeout
		delegate.OnTimeout(err)
	}

	// Success Callback
	successCallback := func(envelope *rony.MessageEnvelope) {
		b, _ := envelope.Marshal()
		delegate.OnComplete(b)
	}

	serverForce := delegate.Flags()&riversdk.RequestServerForced != 0

	// If this request must be sent to the server then executeRemoteCommand
	if serverForce {
		executeRemoteCommand(teamID, teamAccess, r, uint64(requestID), constructor, commandBytes, timeoutCallback, successCallback)
		return
	}

	// If the constructor is a local command then
	handler, ok := r.localCommands[constructor]
	if ok {
		executeLocalCommand(teamID, teamAccess, handler, uint64(requestID), constructor, commandBytes, timeoutCallback, successCallback)
		return
	}

	// If we reached here, then execute the remote commands
	executeRemoteCommand(teamID, teamAccess, r, uint64(requestID), constructor, commandBytes, timeoutCallback, successCallback)

	return
}
func executeLocalCommand(
	teamID int64, teamAccess uint64,
	handler domain.LocalMessageHandler,
	requestID uint64, constructor int64, commandBytes []byte,
	timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler,
) {
	logs.Debug("We execute local command",
		zap.String("C", registry.ConstructorName(constructor)),
	)

	in := &rony.MessageEnvelope{
		Header:      domain.TeamHeader(teamID, teamAccess),
		Constructor: constructor,
		Message:     commandBytes,
		RequestID:   requestID,
	}
	out := &rony.MessageEnvelope{
		Header:    domain.TeamHeader(teamID, teamAccess),
		RequestID: requestID,
	}
	handler(in, out, timeoutCB, successCB)
}
func executeRemoteCommand(
	teamID int64, teamAccess uint64,
	r *River,
	requestID uint64, constructor int64, commandBytes []byte,
	timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler,
) {
	logs.Debug("We execute remote command",
		zap.String("C", registry.ConstructorName(constructor)),
	)
	requestID = uint64(domain.SequentialUniqueID())
	req := &rony.MessageEnvelope{
		Header:      domain.TeamHeader(teamID, teamAccess),
		Constructor: constructor,
		RequestID:   requestID,
		Message:     commandBytes,
		Auth:        nil,
	}

	ctx, cf := context.WithTimeout(context.Background(), domain.HttpRequestTimeout)
	defer cf()
	res, err := r.networkCtrl.SendHttp(ctx, req)
	if res == nil {
		res = &rony.MessageEnvelope{}
		if err != nil {
			rony.ErrorMessage(res, requestID, "E100", err.Error())
		} else {
			rony.ErrorMessage(res, requestID, "E100", "Nil Response")
		}
	}

	successCB(res)
}
