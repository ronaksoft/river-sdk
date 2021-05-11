package group

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/repo"
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

func (r *group) groupFull(e *rony.MessageEnvelope) {
	u := new(msg.GroupFull)
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal GroupFull", zap.Error(err))
		return
	}
	logs.Debug("SyncCtrl applies GroupFull",
		zap.Int64("GroupID", u.Group.ID),
	)

	// save GroupSearch
	err = repo.Groups.SaveFull(u)
	if err != nil {
		logs.Error("SyncCtrl couldn't save GroupFull", zap.Error(err))
	}
	err = repo.Groups.Save(u.Group)
	if err != nil {
		logs.Error("SyncCtrl couldn't save GroupFull's Group", zap.Error(err))
	}

	// Save Users, and notify settings
	_ = repo.Users.Save(u.Users...)
	_ = repo.Groups.SavePhotoGallery(u.Group.ID, u.PhotoGallery...)
	_ = repo.Dialogs.UpdateNotifySetting(u.Group.TeamID, u.Group.ID, int32(msg.PeerType_PeerGroup), u.NotifySettings)
}
