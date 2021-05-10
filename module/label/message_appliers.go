package label

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
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

func (r *label) labelsMany(e *rony.MessageEnvelope) {
	u := &msg.LabelsMany{}
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal LabelsMany", zap.Error(err))
		return
	}

	logs.Debug("SyncCtrl applies LabelsMany", zap.Any("TeamID", e.Get("TeamID", "0")))

	err = repo.Labels.Save(domain.GetTeamID(e), u.Labels...)
	logs.WarnOnErr("SyncCtrl got error on applying LabelsMany", err)

	return
}

func (r *label) labelItems(e *rony.MessageEnvelope) {
	u := &msg.LabelItems{}
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal LabelItems", zap.Error(err))
		return
	}

	logs.Debug("SyncCtrl applies LabelItems")

	_ = repo.Messages.Save(u.Messages...)
	_ = repo.Users.Save(u.Users...)
	_ = repo.Groups.Save(u.Groups...)
}
