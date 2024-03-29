package team

import (
    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/repo"
    "github.com/ronaksoft/river-sdk/internal/request"
    "github.com/ronaksoft/rony"
    "github.com/ronaksoft/rony/errors"
)

/*
   Creation Time: 2021 - May - 10
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func (r *team) teamEdit(da request.Callback) {
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

func (r *team) clientGetTeamCounters(da request.Callback) {
    req := &msg.ClientGetTeamCounters{}
    if err := da.RequestData(req); err != nil {
        return
    }

    unreadCount, mentionCount, err := repo.Dialogs.CountAllUnread(r.SDK().GetConnInfo().PickupUserID(), req.Team.ID, req.WithMutes)

    if err != nil {
        da.Response(rony.C_Error, errors.New("00", err.Error()))
        return
    }

    res := &msg.ClientTeamCounters{
        UnreadCount:  int64(unreadCount),
        MentionCount: int64(mentionCount),
    }

    da.Response(msg.C_ClientTeamCounters, res)
}
