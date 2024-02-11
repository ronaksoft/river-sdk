package team

import (
    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/river-sdk/internal/request"
    "github.com/ronaksoft/river-sdk/module"
)

/*
   Creation Time: 2021 - May - 10
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

type team struct {
    module.Base
}

func New() *team {
    r := &team{}
    r.RegisterHandlers(
        map[int64]request.LocalHandler{
            msg.C_TeamEdit:              r.teamEdit,
            msg.C_ClientGetTeamCounters: r.clientGetTeamCounters,
        },
    )
    r.RegisterUpdateAppliers(
        map[int64]domain.UpdateApplier{
            msg.C_UpdateTeam:              r.updateTeam,
            msg.C_UpdateTeamCreated:       r.updateTeamCreated,
            msg.C_UpdateTeamMemberAdded:   r.updateTeamMemberAdded,
            msg.C_UpdateTeamMemberRemoved: r.updateTeamMemberRemoved,
            msg.C_UpdateTeamMemberStatus:  r.updateTeamMemberStatus,
        },
    )
    r.RegisterMessageAppliers(
        map[int64]domain.MessageApplier{
            msg.C_TeamMembers: r.teamMembers,
            msg.C_TeamsMany:   r.teamsMany,
        },
    )
    return r
}

func (r *team) Name() string {
    return module.Team
}
