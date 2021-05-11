package label

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

func (r *label) updateLabelItemsAdded(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := &msg.UpdateLabelItemsAdded{}
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Debug("SyncCtrl applies UpdateLabelItemsAdded",
		zap.Int64("UpdateID", x.UpdateID),
		zap.Int64s("MsgIDs", x.MessageIDs),
		zap.Int32s("LabelIDs", x.LabelIDs),
	)

	if len(x.MessageIDs) != 0 {
		err := repo.Labels.AddLabelsToMessages(x.LabelIDs, x.TeamID, x.Peer.ID, x.Peer.Type, x.MessageIDs)
		if err != nil {
			return nil, err
		}
		for _, labelID := range x.LabelIDs {
			bar := repo.Labels.GetFilled(x.TeamID, labelID)
			for _, msgID := range x.MessageIDs {
				if msgID > bar.MaxID {
					repo.Labels.Fill(x.TeamID, labelID, bar.MaxID, msgID)
				} else if msgID < bar.MinID {
					repo.Labels.Fill(x.TeamID, labelID, msgID, bar.MinID)
				}
			}
		}

	}

	err = repo.Labels.Save(x.TeamID, x.Labels...)
	if err != nil {
		return nil, err
	}
	return []*msg.UpdateEnvelope{u}, nil
}

func (r *label) updateLabelItemsRemoved(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := &msg.UpdateLabelItemsRemoved{}
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Debug("SyncCtrl applies UpdateLabelItemsRemoved",
		zap.Int64("UpdateID", x.UpdateID),
		zap.Int64s("MsgIDs", x.MessageIDs),
		zap.Int32s("LabelIDs", x.LabelIDs),
		zap.Int64("TeamID", x.TeamID),
	)

	if len(x.MessageIDs) != 0 {
		err := repo.Labels.RemoveLabelsFromMessages(x.LabelIDs, x.TeamID, x.Peer.ID, x.Peer.Type, x.MessageIDs)
		if err != nil {
			return nil, err
		}
	}

	err = repo.Labels.Save(x.TeamID, x.Labels...)
	if err != nil {
		return nil, err
	}
	return []*msg.UpdateEnvelope{u}, nil
}

func (r *label) updateLabelSet(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := &msg.UpdateLabelSet{}
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Debug("SyncCtrl applies UpdateLabelSet",
		zap.Int64("UpdateID", x.UpdateID),
	)

	err = repo.Labels.Set(x.Labels...)
	if err != nil {
		return nil, err
	}
	return []*msg.UpdateEnvelope{u}, nil
}

func (r *label) updateLabelDeleted(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := &msg.UpdateLabelDeleted{}
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Debug("SyncCtrl applies UpdateLabelDeleted",
		zap.Int64("UpdateID", x.UpdateID),
	)

	err = repo.Labels.Delete(x.LabelIDs...)
	if err != nil {
		return nil, err
	}
	return []*msg.UpdateEnvelope{u}, nil
}
