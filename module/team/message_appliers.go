package team

import (
    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/repo"
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
        r.Log().Error("couldn't unmarshal TeamsMany", zap.Error(err))
        return
    }

    err = repo.Users.Save(tm.Users...)
    r.Log().ErrorOnErr("couldn't save teamsMany users", err)

    err = repo.Teams.Clear()
    r.Log().ErrorOnErr("couldn't clear saved teams", err)

    err = repo.Teams.Save(tm.Teams...)
    r.Log().ErrorOnErr("couldn't save teamsMany teams", err)
}

func (r *team) teamMembers(e *rony.MessageEnvelope) {
    tm := &msg.TeamMembers{}
    err := tm.Unmarshal(e.Message)
    if err != nil {
        r.Log().Error("couldn't unmarshal TeamMembers", zap.Error(err))
        return
    }

    err = repo.Users.Save(tm.Users...)
    r.Log().ErrorOnErr("couldn't save teamMembers users", err)
}
