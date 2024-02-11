package team

import (
    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/river-sdk/internal/repo"
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

func (r *team) updateTeamMemberAdded(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
    x := &msg.UpdateTeamMemberAdded{}
    err := x.Unmarshal(u.Update)
    if err != nil {
        return nil, err
    }

    r.Log().Debug("applies UpdateTeamMemberAdded",
        zap.Int64("UpdateID", x.UpdateID),
    )

    _ = repo.Users.Save(x.User)
    _ = repo.Users.SaveContact(x.TeamID, x.Contact)
    err = repo.System.SaveInt(domain.GetContactsGetHashKey(x.TeamID), uint64(x.Hash))
    if err != nil {
        return nil, err
    }
    return []*msg.UpdateEnvelope{u}, nil
}

func (r *team) updateTeamMemberRemoved(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
    x := &msg.UpdateTeamMemberRemoved{}
    err := x.Unmarshal(u.Update)
    if err != nil {
        return nil, err
    }

    r.Log().Debug("applies UpdateTeamMemberRemoved",
        zap.Int64("UpdateID", x.UpdateID),
    )

    if x.UserID == r.SDK().GetConnInfo().PickupUserID() {
        // we have been deleted from the team
        _ = repo.Teams.Delete(x.TeamID)
        _ = repo.Users.DeleteAllContacts(x.TeamID)
    } else {
        _ = repo.Users.DeleteContact(x.TeamID, x.UserID)
        err = repo.System.SaveInt(domain.GetContactsGetHashKey(x.TeamID), uint64(x.Hash))
        if err != nil {
            return nil, err
        }
    }

    return []*msg.UpdateEnvelope{u}, nil
}

// TODO:: improve applier to update data locally
func (r *team) updateTeamMemberStatus(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
    x := &msg.UpdateTeamMemberStatus{}
    err := x.Unmarshal(u.Update)
    if err != nil {
        return nil, err
    }

    r.Log().Debug("applies UpdateTeamMemberStatus",
        zap.Int64("UpdateID", x.UpdateID),
    )

    return []*msg.UpdateEnvelope{u}, nil
}

func (r *team) updateTeamCreated(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
    x := &msg.UpdateTeamCreated{}
    err := x.Unmarshal(u.Update)
    if err != nil {
        return nil, err
    }

    r.Log().Debug("applies UpdateTeamCreated",
        zap.Int64("UpdateID", x.UpdateID),
    )

    err = repo.Teams.Save(x.Team)
    if err != nil {
        return nil, err
    }

    return []*msg.UpdateEnvelope{u}, nil
}

func (r *team) updateTeam(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
    x := &msg.UpdateTeam{}
    err := x.Unmarshal(u.Update)
    if err != nil {
        return nil, err
    }

    r.Log().Debug("applies UpdateTeam",
        zap.Int64("UpdateID", x.UpdateID),
        zap.String("Name", x.Name),
    )

    t, err := repo.Teams.Get(x.TeamID)
    if err != nil {
        return nil, nil
    }

    t.Name = x.Name

    err = repo.Teams.Save(t)
    if err != nil {
        return nil, err
    }

    return []*msg.UpdateEnvelope{u}, nil
}
