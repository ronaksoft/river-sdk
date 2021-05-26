package label

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/request"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/registry"
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

func (r *label) labelsGet(da request.Callback) {
	req := &msg.LabelsGet{}
	if err := da.RequestData(req); err != nil {
		return
	}

	r.Log().Info("LabelGet", zap.Int64("TeamID", da.TeamID()))
	labels, _ := repo.Labels.GetAll(da.TeamID())
	sort.Slice(labels, func(i, j int) bool {
		return labels[i].Count > labels[j].Count
	})
	if len(labels) != 0 {
		r.Log().Debug("found labels locally", zap.Int("L", len(labels)))
		res := &msg.LabelsMany{}
		res.Labels = labels
		da.Response(msg.C_LabelsMany, res)
		return
	}

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *label) labelsDelete(da request.Callback) {
	req := &msg.LabelsDelete{}
	if err := da.RequestData(req); err != nil {
		return
	}

	r.Log().Info("LabelsDelete", zap.Int64("TeamID", da.TeamID()))
	err := repo.Labels.Delete(req.LabelIDs...)

	r.Log().ErrorOnErr("LabelsDelete", err)

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *label) labelsListItems(da request.Callback) {
	req := &msg.LabelsListItems{}
	if err := da.RequestData(req); err != nil {
		return
	}

	// Offline mode
	if !r.SDK().NetCtrl().Connected() {
		r.Log().Debug("are offline then load from local db",
			zap.Int32("LabelID", req.LabelID),
			zap.Int64("MinID", req.MinID),
			zap.Int64("MaxID", req.MaxID),
		)
		messages, users, groups := repo.Labels.ListMessages(req.LabelID, da.TeamID(), req.Limit, req.MinID, req.MaxID)
		fillLabelItems(da, messages, users, groups)
		return
	}

	da.ReplaceCompleteCB(func(m *rony.MessageEnvelope) {
		switch m.Constructor {
		case msg.C_LabelItems:
			x := &msg.LabelItems{}
			err := x.Unmarshal(m.Message)
			r.Log().WarnOnErr("Error On Unmarshal LabelItems", err)

			// 1st sort the received messages by id
			sort.Slice(x.Messages, func(i, j int) bool {
				return x.Messages[i].ID > x.Messages[j].ID
			})

			// Fill Messages Hole
			if msgCount := len(x.Messages); msgCount > 0 {
				r.Log().Debug("Update Label Range",
					zap.Int32("LabelID", x.LabelID),
					zap.Int64("MinID", x.Messages[msgCount-1].ID),
					zap.Int64("MaxID", x.Messages[0].ID),
				)

				switch {
				case req.MinID == 0 && req.MaxID != 0:
					_ = repo.Labels.Fill(da.TeamID(), req.LabelID, x.Messages[msgCount-1].ID, req.MaxID)
				case req.MinID != 0 && req.MaxID == 0:
					_ = repo.Labels.Fill(da.TeamID(), req.LabelID, req.MinID, x.Messages[0].ID)
				case req.MinID == 0 && req.MaxID == 0:
					_ = repo.Labels.Fill(da.TeamID(), req.LabelID, x.Messages[msgCount-1].ID, x.Messages[0].ID)
				}
			}
		default:
			r.Log().Warn("LabelModule received unexpected response", zap.String("C", registry.ConstructorName(m.Constructor)))
		}
	})
	switch {
	case req.MinID == 0 && req.MaxID == 0:
		r.SDK().QueueCtrl().EnqueueCommand(da)
	case req.MinID == 0 && req.MaxID != 0:
		b, _ := repo.Labels.GetLowerFilled(da.TeamID(), req.LabelID, req.MaxID)
		if !b {
			r.Log().Info("LabelModule detected label hole (With MaxID Only)",
				zap.Int32("LabelID", req.LabelID),
				zap.Int64("MaxID", req.MaxID),
				zap.Int64("MinID", req.MinID),
			)
			r.SDK().QueueCtrl().EnqueueCommand(da)
			return
		}
		messages, users, groups := repo.Labels.ListMessages(req.LabelID, da.TeamID(), req.Limit, 0, req.MaxID)
		fillLabelItems(da, messages, users, groups)
	case req.MinID != 0 && req.MaxID == 0:
		b, _ := repo.Labels.GetUpperFilled(da.TeamID(), req.LabelID, req.MinID)
		if !b {
			r.Log().Info("detected label hole (With MinID Only)",
				zap.Int32("LabelID", req.LabelID),
				zap.Int64("MinID", req.MinID),
				zap.Int64("MaxID", req.MaxID),
			)
			r.SDK().QueueCtrl().EnqueueCommand(da)
			return
		}
		messages, users, groups := repo.Labels.ListMessages(req.LabelID, da.TeamID(), req.Limit, req.MinID, 0)
		fillLabelItems(da, messages, users, groups)
	default:
		r.SDK().QueueCtrl().EnqueueCommand(da)
		return
	}
}

func (r *label) labelAddToMessage(da request.Callback) {
	req := &msg.LabelsAddToMessage{}
	if err := da.RequestData(req); err != nil {
		return
	}

	r.Log().Debug("LabelsAddToMessage local handler called",
		zap.Int64s("MsgIDs", req.MessageIDs),
		zap.Int32s("LabelIDs", req.LabelIDs),
	)
	if len(req.MessageIDs) != 0 {
		_ = repo.Labels.AddLabelsToMessages(req.LabelIDs, da.TeamID(), req.Peer.ID, int32(req.Peer.Type), req.MessageIDs)
		for _, labelID := range req.LabelIDs {
			bar := repo.Labels.GetFilled(da.TeamID(), labelID)
			for _, msgID := range req.MessageIDs {
				if msgID > bar.MaxID {
					_ = repo.Labels.Fill(da.TeamID(), labelID, bar.MaxID, msgID)
				} else if msgID < bar.MinID {
					_ = repo.Labels.Fill(da.TeamID(), labelID, msgID, bar.MinID)
				}
			}
		}
	}

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)

}

func (r *label) labelRemoveFromMessage(da request.Callback) {
	req := &msg.LabelsRemoveFromMessage{}
	if err := da.RequestData(req); err != nil {
		return
	}

	r.Log().Debug("LabelsRemoveFromMessage local handler called",
		zap.Int64s("MsgIDs", req.MessageIDs),
		zap.Int32s("LabelIDs", req.LabelIDs),
	)

	if len(req.MessageIDs) != 0 {
		_ = repo.Labels.RemoveLabelsFromMessages(req.LabelIDs, da.TeamID(), req.Peer.ID, int32(req.Peer.Type), req.MessageIDs)
	}

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)

}

func fillLabelItems(da request.Callback, messages []*msg.UserMessage, users []*msg.User, groups []*msg.Group) {
	res := new(msg.LabelItems)
	res.Messages = messages
	res.Users = users
	res.Groups = groups

	da.Response(msg.C_LabelItems, res)
}
