package user

import (
	"git.ronaksoft.com/river/msg/go/msg"
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

func (r *user) updateUsername(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateUsername)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	r.Log().Debug("UserModule applies UpdateUsername",
		zap.Int64("UpdateID", x.UpdateID),
	)

	if x.UserID == r.SDK().SyncCtrl().GetUserID() {
		r.SDK().GetConnInfo().ChangeUserID(x.UserID)
		r.SDK().GetConnInfo().ChangeUsername(x.Username)
		r.SDK().GetConnInfo().ChangeFirstName(x.FirstName)
		r.SDK().GetConnInfo().ChangeLastName(x.LastName)
		r.SDK().GetConnInfo().ChangePhone(x.Phone)
		r.SDK().GetConnInfo().ChangeBio(x.Bio)
		r.SDK().GetConnInfo().Save()
	}

	err = repo.Users.UpdateProfile(x.UserID, x.FirstName, x.LastName, x.Username, x.Bio, x.Phone)
	if err != nil {
		return nil, err
	}

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (r *user) updateUserPhoto(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateUserPhoto)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	r.Log().Debug("UserModule applies UpdateUserPhoto",
		zap.Int64("UpdateID", x.UpdateID),
		zap.Any("PhotoID", x.PhotoID),
	)

	if x.Photo != nil {
		err = repo.Users.UpdatePhoto(x.UserID, x.Photo)
		if err != nil {
			r.Log().Warn("UserModule got error on updating user's profile photo",
				zap.Int64("UserID", x.UserID),
				zap.Any("Photo", x.Photo),
			)
		}
	}

	for _, photoID := range x.DeletedPhotoIDs {
		_ = repo.Users.RemovePhotoGallery(x.UserID, photoID)
	}

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (r *user) updateUserBlocked(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := &msg.UpdateUserBlocked{}
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	r.Log().Debug("UserModule applies UpdateUserBlocked",
		zap.Int64("UpdateID", x.UpdateID),
	)

	err = repo.Users.UpdateBlocked(x.UserID, x.Blocked)
	if err != nil {
		return nil, err
	}
	return []*msg.UpdateEnvelope{u}, nil
}
