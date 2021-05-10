package team

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"github.com/ronaksoft/rony"
)

/*
   Creation Time: 2021 - May - 10
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func (r *team) teamEdit(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.TeamEdit{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	team, _ := repo.Teams.Get(req.TeamID)

	if team != nil {
		team.Name = req.Name
		_ = repo.Teams.Save(team)
	}

	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}
