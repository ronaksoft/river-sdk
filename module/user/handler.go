package user

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/uiexec"
	"github.com/ronaksoft/rony"
	"go.uber.org/zap"
	"sort"
)

/*
   Creation Time: 2021 - May - 10
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func (r *user) usersGetFull(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.UsersGetFull{}
	if err := req.Unmarshal(in.Message); err != nil {
		r.Log().Error("UserModule::usersGetFull()-> Unmarshal()", zap.Error(err))
		return
	}
	userIDs := domain.MInt64B{}
	for _, v := range req.Users {
		userIDs[v.UserID] = true
	}

	users, outDated, _ := repo.Users.GetManyWithOutdated(userIDs.ToArray())
	allResolved := len(users) == len(userIDs)
	if allResolved {
		res := &msg.UsersMany{}
		for _, user := range users {
			user.PhotoGallery = repo.Users.GetPhotoGallery(user.ID)
			sort.Slice(user.PhotoGallery, func(i, j int) bool {
				return user.PhotoGallery[i].PhotoID > user.PhotoGallery[j].PhotoID
			})
		}
		res.Users = users

		out.Constructor = msg.C_UsersMany
		out.Message, _ = res.Marshal()
		uiexec.ExecSuccessCB(da.OnComplete, out)

		if len(outDated) > 0 {
			req.Users = req.Users[:0]
			for _, user := range outDated {
				req.Users = append(req.Users, &msg.InputUser{
					UserID:     user.ID,
					AccessHash: user.AccessHash,
				})
			}
			in.Fill(in.RequestID, in.Constructor, req, in.Header...)
			r.SDK().QueueCtrl().EnqueueCommand(in, nil, nil, da.UI())
		}
		return
	}

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(in, da.OnTimeout, da.OnComplete, da.UI())
}

func (r *user) usersGet(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.UsersGet{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}
	userIDs := domain.MInt64B{}
	for _, v := range req.Users {
		userIDs[v.UserID] = true
	}

	users, outDated, _ := repo.Users.GetManyWithOutdated(userIDs.ToArray())
	allResolved := len(users) == len(userIDs)
	if allResolved {
		res := new(msg.UsersMany)
		res.Users = users

		out.Constructor = msg.C_UsersMany
		out.Message, _ = res.Marshal()
		uiexec.ExecSuccessCB(da.OnComplete, out)

		if len(outDated) > 0 {
			req.Users = req.Users[:0]
			for _, user := range outDated {
				req.Users = append(req.Users, &msg.InputUser{
					UserID:     user.ID,
					AccessHash: user.AccessHash,
				})
			}
			in.Fill(in.RequestID, in.Constructor, req, in.Header...)
			r.SDK().QueueCtrl().EnqueueCommand(in, nil, nil, da.UI())
		}
		return
	}

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(in, da.OnTimeout, da.OnComplete, da.UI())
}
