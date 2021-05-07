package mini

import (
	"context"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/minirepo"
	"github.com/ronaksoft/rony"
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

	logs.Info("MiniRiver Starting")
	logs.SetSentry(r.ConnInfo.AuthID, r.ConnInfo.UserID, r.sentryDSN)

	minirepo.MustInit(r.dbPath)

	// Update Authorizations
	r.network.SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey)

	// Start Controllers
	r.network.Start()

	domain.StartTime = time.Now()
	domain.WindowLog = func(txt string) {}
	logs.Info("MiniRiver Started")

	err := r.syncServerTime()
	if err != nil {
		logs.Warn("MiniRiver got error on get server time", zap.Error(err))
	}

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

	serverForce := delegate.Flags()&RequestServerForced != 0
	rda := NewDelegateAdapter(delegate)

	// If this request must be sent to the server then executeRemoteCommand
	if serverForce {
		executeRemoteCommand(teamID, teamAccess, r, uint64(requestID), constructor, commandBytes, rda)
		return
	}

	// If the constructor is a local command then
	handler, ok := r.localCommands[constructor]
	if ok {
		executeLocalCommand(teamID, teamAccess, handler, uint64(requestID), constructor, commandBytes, rda)
		return
	}

	// If we reached here, then execute the remote commands
	executeRemoteCommand(teamID, teamAccess, r, uint64(requestID), constructor, commandBytes, rda)

	return
}
func executeLocalCommand(
	teamID int64, teamAccess uint64,
	handler LocalHandler,
	requestID uint64, constructor int64, commandBytes []byte,
	da *DelegateAdapter,
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
	handler(in, out, da)
}
func executeRemoteCommand(
	teamID int64, teamAccess uint64,
	r *River,
	requestID uint64, constructor int64, commandBytes []byte,
	da *DelegateAdapter,
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
	res, err := r.network.SendHttp(ctx, req)
	if res == nil {
		res = &rony.MessageEnvelope{}
		if err != nil {
			rony.ErrorMessage(res, requestID, "E100", err.Error())
		} else {
			rony.ErrorMessage(res, requestID, "E100", "Nil Response")
		}
	}

	da.OnComplete(res)
}

func (r *River) SetTeam(teamID int64, teamAccessHash int64) {
	domain.SetCurrentTeam(teamID, uint64(teamAccessHash))
}
