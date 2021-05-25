package mini

import (
	"context"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/minirepo"
	"git.ronaksoft.com/river/sdk/internal/request"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/errors"
	"github.com/ronaksoft/rony/registry"
	"go.uber.org/zap"
	"runtime"
	"sync"
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

	if r.getLastUpdateID() == 0 {
		// run in sync for the first time
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			r.syncContacts()
			wg.Done()
		}()
		go func() {
			r.syncDialogs()
			wg.Done()
		}()
		wg.Wait()
	} else {
		// run in background
		go r.syncContacts()
		go r.syncDialogs()
	}

	return nil
}

// ExecuteCommand is a wrapper function to pass the request to the queueController, to be passed to networkController for final
// delivery to the server. SDK uses GetCurrentTeam() to detect the targeted team of the request
func (r *River) ExecuteCommand(constructor int64, commandBytes []byte, delegate RequestDelegate) (requestID int64, err error) {
	return r.executeCommand(domain.GetCurrTeamID(), domain.GetCurrTeamAccess(), constructor, commandBytes, delegate)
}

// ExecuteCommandWithTeam is similar to ExecuteTeam but explicitly defines the target team
func (r *River) ExecuteCommandWithTeam(teamID, accessHash, constructor int64, commandBytes []byte, delegate RequestDelegate) (requestID int64, err error) {
	return r.executeCommand(teamID, uint64(accessHash), constructor, commandBytes, delegate)
}

func (r *River) executeCommand(
	teamID int64, teamAccess uint64, constructor int64, commandBytes []byte, delegate RequestDelegate,
) (requestID int64, err error) {
	if registry.ConstructorName(constructor) == "" {
		return 0, domain.ErrInvalidConstructor
	}

	requestID = int64(domain.NextRequestID())
	serverForce := delegate.Flags()&request.ServerForced != 0
	rda := request.DelegateAdapter(uint64(requestID), constructor, delegate, true)

	// If this request must be sent to the server then executeRemoteCommand
	if serverForce {
		r.executeRemoteCommand(teamID, teamAccess,  commandBytes, rda)
		return
	}

	// If the constructor is a local command then
	handler, ok := r.localCommands[constructor]
	if ok {
		r.executeLocalCommand(teamID, teamAccess, handler, commandBytes, rda)
		return
	}

	// If we reached here, then execute the remote commands
	r.executeRemoteCommand(teamID, teamAccess, commandBytes, rda)

	return
}
func (r *River) executeLocalCommand(
	teamID int64, teamAccess uint64,
	handler request.LocalHandler,
	commandBytes []byte,
	da request.Callback,
) {
	logger.Debug("execute local command",
		zap.String("C", registry.ConstructorName(da.Constructor())),
	)

	in := &rony.MessageEnvelope{
		Header:      domain.TeamHeader(teamID, teamAccess),
		Constructor: da.Constructor(),
		Message:     commandBytes,
		RequestID:   da.RequestID(),
	}
	out := &rony.MessageEnvelope{
		Header:    domain.TeamHeader(teamID, teamAccess),
		RequestID: da.RequestID(),
	}
	handler(in, out, da)
}
func (r *River) executeRemoteCommand(
	teamID int64, teamAccess uint64, commandBytes []byte,
	da request.Callback,
) {
	logger.Debug("execute remote command",
		zap.String("C", registry.ConstructorName(da.Constructor())),
	)

	req := &rony.MessageEnvelope{
		Header:      domain.TeamHeader(teamID, teamAccess),
		Constructor: da.Constructor(),
		RequestID:   da.RequestID(),
		Message:     commandBytes,
		Auth:        nil,
	}

	ctx, cf := context.WithTimeout(context.Background(), domain.HttpRequestTimeout)
	defer cf()
	res, err := r.network.SendHttp(ctx, req)
	if res == nil {
		res = &rony.MessageEnvelope{}
		if err != nil {
			errors.New("E100", err.Error()).ToEnvelope(res)
		} else {
			errors.New("E100", "Nil Response").ToEnvelope(res)
		}
	}

	da.OnComplete(res)
}

func (r *River) SetTeam(teamID int64, teamAccessHash int64) {
	domain.SetCurrentTeam(teamID, uint64(teamAccessHash))
}
