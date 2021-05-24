package group

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/repo"
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

func (r *group) updateGroupParticipantAdmin(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := &msg.UpdateGroupParticipantAdmin{}
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	r.Log().Debug("GroupModule applies UpdateGroupParticipantAdmin",
		zap.Int64("UpdateID", x.UpdateID),
	)

	res := []*msg.UpdateEnvelope{u}
	repo.Groups.UpdateMemberType(x.GroupID, x.UserID, x.IsAdmin)
	return res, nil
}

func (r *group) updateGroupPhoto(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateGroupPhoto)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	r.Log().Debug("GroupModule applies UpdateGroupPhoto",
		zap.Int64("GroupID", x.GroupID),
		zap.Int64("UpdateID", x.UpdateID),
		zap.Int64("PhotoID", x.PhotoID),
	)

	if x.Photo != nil {
		_ = repo.Groups.UpdatePhoto(x.GroupID, x.Photo)
	}

	if x.PhotoID != 0 {
		if x.Photo != nil && x.Photo.PhotoSmall.FileID != 0 {
			_ = repo.Groups.SavePhotoGallery(x.GroupID, x.Photo)
		} else {
			_ = repo.Groups.RemovePhotoGallery(x.GroupID, x.PhotoID)
		}
	}

	groupFull, _ := repo.Groups.GetFull(x.GroupID)
	if groupFull != nil {
		group, _ := repo.Groups.Get(x.GroupID)
		if group != nil {
			groupFull.Group = group
			_ = repo.Groups.SaveFull(groupFull)
		}
	}

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (r *group) updateGroupAdmins(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := &msg.UpdateGroupAdmins{}
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	r.Log().Debug("GroupModule applies UpdateGroupAdmins",
		zap.Int64("GroupID", x.GroupID),
		zap.Int64("UpdateID", x.UpdateID),
	)

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (r *group) updateGroupAdminOnly(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := &msg.UpdateGroupAdminOnly{}
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	r.Log().Debug("GroupModule applies UpdateGroupAdminOnly",
		zap.Int64("GroupID", x.GroupID),
		zap.Int64("UpdateID", x.UpdateID),
	)

	group, _ := repo.Groups.Get(x.GroupID)
	if group != nil {
		dialog, _ := repo.Dialogs.Get(group.TeamID, group.ID, int32(msg.PeerType_PeerGroup))
		if dialog != nil {
			dialog.ReadOnly = repo.Groups.HasFlag(group.Flags, msg.GroupFlags_GroupFlagsAdminOnly) && !repo.Groups.HasFlag(group.Flags, msg.GroupFlags_GroupFlagsAdmin)
			_ = repo.Dialogs.Save(dialog)
		}
	}

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}
