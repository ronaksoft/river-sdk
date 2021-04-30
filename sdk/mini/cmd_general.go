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

func (r *River) ExecuteCommand(
	teamID int64, teamAccess uint64, constructor int64, commandBytes []byte, delegate riversdk.RequestDelegate,
) (requestID int64, err error) {
	if registry.ConstructorName(constructor) == "" {
		return 0, domain.ErrInvalidConstructor
	}
	requestID = domain.SequentialUniqueID()
	req := &rony.MessageEnvelope{
		Header:      domain.TeamHeader(teamID, teamAccess),
		Constructor: constructor,
		RequestID:   uint64(requestID),
		Message:     commandBytes,
		Auth:        nil,
	}

	ctx, cf := context.WithTimeout(context.Background(), domain.HttpRequestTimeout)
	defer cf()
	res, err := r.networkCtrl.SendHttp(ctx, req)
	if res == nil {
		res = &rony.MessageEnvelope{}
		if err != nil {
			rony.ErrorMessage(res, uint64(requestID), "E100", err.Error())
		} else {
			rony.ErrorMessage(res, uint64(requestID), "E100", "Nil Response")
		}
	}
	resBytes, _ := res.Marshal()
	delegate.OnComplete(resBytes)
	return
}

func (r *River) MessagesGetDialogs() {}
func (r *River) MessagesSendMedia()  {}
func (r *River) MessagesSend()       {}
