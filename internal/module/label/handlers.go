package label

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/uiexec"
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

func (r *label) labelsGet(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.LabelsGet{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	logs.Info("LabelGet", zap.Int64("TeamID", domain.GetTeamID(in)))
	labels := repo.Labels.GetAll(domain.GetTeamID(in))
	sort.Slice(labels, func(i, j int) bool {
		return labels[i].Count > labels[j].Count
	})
	if len(labels) != 0 {
		logs.Debug("We found labels locally", zap.Int("L", len(labels)))
		res := &msg.LabelsMany{}
		res.Labels = labels

		out.Constructor = msg.C_LabelsMany
		out.Message, _ = res.Marshal()
		uiexec.ExecSuccessCB(successCB, out)
		return
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *label) labelsDelete(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.LabelsDelete{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	logs.Info("LabelsDelete", zap.Int64("TeamID", domain.GetTeamID(in)))
	err := repo.Labels.Delete(req.LabelIDs...)

	logs.ErrorOnErr("LabelsDelete", err)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *label) labelsListItems(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.LabelsListItems{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	// Offline mode
	if !r.networkCtrl.Connected() {
		logs.Debug("We are offline then load from local db",
			zap.Int32("LabelID", req.LabelID),
			zap.Int64("MinID", req.MinID),
			zap.Int64("MaxID", req.MaxID),
		)
		messages, users, groups := repo.Labels.ListMessages(req.LabelID, domain.GetTeamID(in), req.Limit, req.MinID, req.MaxID)
		fillLabelItems(out, messages, users, groups, in.RequestID, successCB)
		return
	}

	preSuccessCB := func(m *rony.MessageEnvelope) {
		switch m.Constructor {
		case msg.C_LabelItems:
			x := &msg.LabelItems{}
			err := x.Unmarshal(m.Message)
			logs.WarnOnErr("Error On Unmarshal LabelItems", err)

			// 1st sort the received messages by id
			sort.Slice(x.Messages, func(i, j int) bool {
				return x.Messages[i].ID > x.Messages[j].ID
			})

			// Fill Messages Hole
			if msgCount := len(x.Messages); msgCount > 0 {
				logs.Debug("Update Label Range",
					zap.Int32("LabelID", x.LabelID),
					zap.Int64("MinID", x.Messages[msgCount-1].ID),
					zap.Int64("MaxID", x.Messages[0].ID),
				)

				switch {
				case req.MinID == 0 && req.MaxID != 0:
					_ = repo.Labels.Fill(domain.GetTeamID(in), req.LabelID, x.Messages[msgCount-1].ID, req.MaxID)
				case req.MinID != 0 && req.MaxID == 0:
					_ = repo.Labels.Fill(domain.GetTeamID(in), req.LabelID, req.MinID, x.Messages[0].ID)
				case req.MinID == 0 && req.MaxID == 0:
					_ = repo.Labels.Fill(domain.GetTeamID(in), req.LabelID, x.Messages[msgCount-1].ID, x.Messages[0].ID)
				}
			}
		default:
			logs.Warn("We received unexpected response", zap.String("C", registry.ConstructorName(m.Constructor)))
		}

		successCB(m)
	}

	switch {
	case req.MinID == 0 && req.MaxID == 0:
		r.queueCtrl.EnqueueCommand(in, timeoutCB, preSuccessCB, true)
	case req.MinID == 0 && req.MaxID != 0:
		b, _ := repo.Labels.GetLowerFilled(domain.GetTeamID(in), req.LabelID, req.MaxID)
		if !b {
			logs.Info("River detected label hole (With MaxID Only)",
				zap.Int32("LabelID", req.LabelID),
				zap.Int64("MaxID", req.MaxID),
				zap.Int64("MinID", req.MinID),
			)
			r.queueCtrl.EnqueueCommand(in, timeoutCB, preSuccessCB, true)
			return
		}
		messages, users, groups := repo.Labels.ListMessages(req.LabelID, domain.GetTeamID(in), req.Limit, 0, req.MaxID)
		fillLabelItems(out, messages, users, groups, in.RequestID, preSuccessCB)
	case req.MinID != 0 && req.MaxID == 0:
		b, _ := repo.Labels.GetUpperFilled(domain.GetTeamID(in), req.LabelID, req.MinID)
		if !b {
			logs.Info("River detected label hole (With MinID Only)",
				zap.Int32("LabelID", req.LabelID),
				zap.Int64("MinID", req.MinID),
				zap.Int64("MaxID", req.MaxID),
			)
			r.queueCtrl.EnqueueCommand(in, timeoutCB, preSuccessCB, true)
			return
		}
		messages, users, groups := repo.Labels.ListMessages(req.LabelID, domain.GetTeamID(in), req.Limit, req.MinID, 0)
		fillLabelItems(out, messages, users, groups, in.RequestID, preSuccessCB)
	default:
		r.queueCtrl.EnqueueCommand(in, timeoutCB, preSuccessCB, true)
		return
	}
}

func (r *label) labelAddToMessage(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.LabelsAddToMessage{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	logs.Debug("LabelsAddToMessage local handler called",
		zap.Int64s("MsgIDs", req.MessageIDs),
		zap.Int32s("LabelIDs", req.LabelIDs),
	)
	if len(req.MessageIDs) != 0 {
		_ = repo.Labels.AddLabelsToMessages(req.LabelIDs, domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.MessageIDs)
		for _, labelID := range req.LabelIDs {
			bar := repo.Labels.GetFilled(domain.GetTeamID(in), labelID)
			for _, msgID := range req.MessageIDs {
				if msgID > bar.MaxID {
					_ = repo.Labels.Fill(domain.GetTeamID(in), labelID, bar.MaxID, msgID)
				} else if msgID < bar.MinID {
					_ = repo.Labels.Fill(domain.GetTeamID(in), labelID, msgID, bar.MinID)
				}
			}
		}
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)

}

func (r *label) labelRemoveFromMessage(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.LabelsRemoveFromMessage{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	logs.Debug("LabelsRemoveFromMessage local handler called",
		zap.Int64s("MsgIDs", req.MessageIDs),
		zap.Int32s("LabelIDs", req.LabelIDs),
	)

	if len(req.MessageIDs) != 0 {
		_ = repo.Labels.RemoveLabelsFromMessages(req.LabelIDs, domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.MessageIDs)
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)

}

func fillLabelItems(out *rony.MessageEnvelope, messages []*msg.UserMessage, users []*msg.User, groups []*msg.Group, requestID uint64, successCB domain.MessageHandler) {
	res := new(msg.LabelItems)
	res.Messages = messages
	res.Users = users
	res.Groups = groups

	out.RequestID = requestID
	out.Constructor = msg.C_LabelItems
	out.Message, _ = res.Marshal()
	uiexec.ExecSuccessCB(successCB, out)
}
