package user

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

func (r *user) usersMany(e *rony.MessageEnvelope) {
	x := new(msg.UsersMany)
	err := x.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal UsersMany", zap.Error(err))
		return
	}
	logs.Debug("SyncCtrl applies usersMany",
		zap.Int("Users", len(x.Users)),
	)
	_ = repo.Users.Save(x.Users...)
}
