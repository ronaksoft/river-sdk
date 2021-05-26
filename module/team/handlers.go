package team

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/request"
	"git.ronaksoft.com/river/sdk/internal/uiexec"
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

func (r *team) teamEdit(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.TeamEdit{}
	if err := da.RequestData(req); err != nil {
		return
	}

	team, _ := repo.Teams.Get(req.TeamID)

	if team != nil {
		team.Name = req.Name
		_ = repo.Teams.Save(team)
	}

	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *team) clientGetTeamCounters(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.ClientGetTeamCounters{}
	if err := da.RequestData(req); err != nil {
		return
	}

	unreadCount, mentionCount, err := repo.Dialogs.CountAllUnread(r.SDK().GetConnInfo().PickupUserID(), req.Team.ID, req.WithMutes)

	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	res := &msg.ClientTeamCounters{
		UnreadCount:  int64(unreadCount),
		MentionCount: int64(mentionCount),
	}

	out.Fill(in.RequestID, msg.C_ClientTeamCounters, res)
	uiexec.ExecSuccessCB(da.OnComplete, out)
}
