package user

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/request"
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

func (r *user) usersGetFull(da request.Callback) {
	req := &msg.UsersGetFull{}
	if err := da.RequestData(req); err != nil {
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
		da.Response(msg.C_UsersMany, res)

		if len(outDated) > 0 {
			req.Users = req.Users[:0]
			for _, user := range outDated {
				req.Users = append(req.Users, &msg.InputUser{
					UserID:     user.ID,
					AccessHash: user.AccessHash,
				})
			}
			// TODO:: update da with new request
			r.SDK().QueueCtrl().EnqueueCommand(da)
		}
		return
	}

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *user) usersGet(da request.Callback) {
	req := &msg.UsersGet{}
	if err := da.RequestData(req); err != nil {
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
		da.Response(msg.C_UsersMany, res)

		if len(outDated) > 0 {
			req.Users = req.Users[:0]
			for _, user := range outDated {
				req.Users = append(req.Users, &msg.InputUser{
					UserID:     user.ID,
					AccessHash: user.AccessHash,
				})
			}
			// TODO:: update da with new request
			r.SDK().QueueCtrl().EnqueueCommand(da)
		}
		return
	}

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)
}
