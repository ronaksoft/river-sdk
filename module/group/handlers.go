package group

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/request"
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

func (r *group) groupsEditTitle(da request.Callback) {
	req := &msg.GroupsEditTitle{}
	if err := da.RequestData(req); err != nil {
		return
	}

	repo.Groups.UpdateTitle(req.GroupID, req.Title)

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)

}

func (r *group) groupAddUser(da request.Callback) {
	req := &msg.GroupsAddUser{}
	if err := da.RequestData(req); err != nil {
		return
	}

	user, _ := repo.Users.Get(req.User.UserID)
	if user != nil {
		gp := &msg.GroupParticipant{
			AccessHash: req.User.AccessHash,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			UserID:     req.User.UserID,
			Type:       msg.ParticipantType_ParticipantTypeMember,
		}
		_ = repo.Groups.AddParticipant(req.GroupID, gp)
	}

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)

}

func (r *group) groupDeleteUser(da request.Callback) {
	req := &msg.GroupsDeleteUser{}
	if err := da.RequestData(req); err != nil {
		return
	}

	err := repo.Groups.RemoveParticipant(req.GroupID, req.User.UserID)
	if err != nil {
		r.Log().Error("got error on GroupDeleteUser local handler", zap.Error(err))
	}

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *group) groupsGetFull(da request.Callback) {
	req := &msg.GroupsGetFull{}
	if err := da.RequestData(req); err != nil {
		return
	}

	res, err := repo.Groups.GetFull(req.GroupID)
	if err != nil {
		r.SDK().QueueCtrl().EnqueueCommand(da)
		return
	}

	// NotifySettings
	dlg, _ := repo.Dialogs.Get(da.TeamID(), req.GroupID, int32(msg.PeerType_PeerGroup))
	if dlg == nil {
		r.SDK().QueueCtrl().EnqueueCommand(da)
		return
	}
	res.NotifySettings = dlg.NotifySettings

	// Get Group PhotoGallery
	res.PhotoGallery, err = repo.Groups.GetPhotoGallery(req.GroupID)
	if err != nil {
		r.Log().Error("got error on GetPhotoGallery in local handler", zap.Error(err))
	}

	// Users
	userIDs := domain.MInt64B{}
	for _, v := range res.Participants {
		userIDs[v.UserID] = true
	}
	users, _ := repo.Users.GetMany(userIDs.ToArray())
	if len(res.Participants) != len(users) {
		r.SDK().QueueCtrl().EnqueueCommand(da)
		return
	}
	res.Users = users
	da.Response(msg.C_GroupFull, res)
}

func (r *group) groupUpdateAdmin(da request.Callback) {
	req := &msg.GroupsUpdateAdmin{}
	if err := da.RequestData(req); err != nil {
		return
	}

	repo.Groups.UpdateMemberType(req.GroupID, req.User.UserID, req.Admin)

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *group) groupToggleAdmin(da request.Callback) {
	req := &msg.GroupsToggleAdmins{}
	if err := da.RequestData(req); err != nil {
		return
	}

	err := repo.Groups.ToggleAdmins(req.GroupID, req.AdminEnabled)
	if err != nil {
		r.Log().Warn("got error on local handler for GroupToggleAdmin", zap.Error(err))
	}

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *group) groupRemovePhoto(da request.Callback) {
	req := &msg.GroupsRemovePhoto{}
	if err := da.RequestData(req); err != nil {
		return
	}

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)

	group, _ := repo.Groups.Get(req.GroupID)
	if group == nil {
		return
	}

	if group.Photo != nil && group.Photo.PhotoID == req.PhotoID {
		_ = repo.Groups.UpdatePhoto(req.GroupID, &msg.GroupPhoto{
			PhotoBig:   &msg.FileLocation{},
			PhotoSmall: &msg.FileLocation{},
			PhotoID:    0,
		})
	}

	_ = repo.Users.RemovePhotoGallery(r.SDK().GetConnInfo().PickupUserID(), req.PhotoID)
}
