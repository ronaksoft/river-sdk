package team

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"github.com/ronaksoft/rony"
	"go.uber.org/zap"
)

/*
   Creation Time: 2021 - May - 10
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func (r *team) teamsMany(e *rony.MessageEnvelope) {
	tm := &msg.TeamsMany{}
	err := tm.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal TeamsMany", zap.Error(err))
		return
	}

	err = repo.Users.Save(tm.Users...)
	logs.ErrorOnErr("SyncCtrl couldn't save teamsMany users", err)

	err = repo.Teams.Clear()
	logs.ErrorOnErr("SyncCtrl couldn't clear saved teams", err)

	err = repo.Teams.Save(tm.Teams...)
	logs.ErrorOnErr("SyncCtrl couldn't save teamsMany teams", err)
}

func (r *team) teamMembers(e *rony.MessageEnvelope) {
	tm := &msg.TeamMembers{}
	err := tm.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal TeamMembers", zap.Error(err))
		return
	}

	err = repo.Users.Save(tm.Users...)
	logs.ErrorOnErr("SyncCtrl couldn't save teamMembers users", err)
}
