package mini

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"github.com/ronaksoft/rony"
	"go.uber.org/zap"
)

/*
   Creation Time: 2021 - May - 01
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func (r *River) messagesSendMedia(in, out *rony.MessageEnvelope, da *DelegateAdapter) {
	req := &msg.MessagesSendMedia{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	req.RandomID = domain.SequentialUniqueID()
	requestBytes, _ := req.Marshal()
	r.networkCtrl.HttpCommand(
		&rony.MessageEnvelope{
			Constructor: msg.C_MessagesSendMedia,
			RequestID:   uint64(req.RandomID),
			Message:     requestBytes,
			Header:      in.Header,
		}, da.OnTimeout, da.OnComplete,
	)
}

func (r *River) messagesGetDialogs(in, out *rony.MessageEnvelope, da *DelegateAdapter) {
	req := &msg.MessagesGetDialogs{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}
	res := &msg.MessagesDialogs{}
	res.Dialogs = repo.Dialogs.List(domain.GetTeamID(in), req.Offset, req.Limit)
	res.Count = repo.Dialogs.CountDialogs(domain.GetTeamID(in))

	// If the localDB had no data send the request to server
	if len(res.Dialogs) == 0 {
		r.networkCtrl.HttpCommand(in, da.OnTimeout, da.OnComplete)
		return
	}

	mUsers := domain.MInt64B{}
	mGroups := domain.MInt64B{}
	mMessages := domain.MInt64B{}
	for _, dialog := range res.Dialogs {
		if dialog.PeerType == int32(msg.PeerType_PeerUser) {
			if dialog.PeerID != 0 {
				mUsers[dialog.PeerID] = true
			}
		}
		mMessages[dialog.TopMessageID] = true
	}

	// Load Messages
	res.Messages = repo.Messages.GetMany(mMessages.ToArray())
	for _, m := range res.Messages {
		switch msg.PeerType(m.PeerType) {
		case msg.PeerType_PeerUser:
			mUsers[m.PeerID] = true
		case msg.PeerType_PeerGroup:
			mGroups[m.PeerID] = true
		}
		if m.SenderID != 0 {
			mUsers[m.SenderID] = true
		}
		if m.FwdSenderID != 0 {
			mUsers[m.FwdSenderID] = true
		}

		// load MessageActionData users
		actUserIDs := domain.ExtractActionUserIDs(m.MessageAction, m.MessageActionData)
		for _, id := range actUserIDs {
			if id != 0 {
				mUsers[id] = true
			}
		}
	}
	res.Groups, _ = repo.Groups.GetMany(mGroups.ToArray())
	if len(res.Groups) != len(mGroups) {
		logs.Warn("River found unmatched dialog groups", zap.Int("Got", len(res.Groups)), zap.Int("Need", len(mGroups)))
		for groupID := range mGroups {
			found := false
			for _, g := range res.Groups {
				if g.ID == groupID {
					found = true
					break
				}
			}
			if !found {
				logs.Warn("missed group", zap.Int64("GroupID", groupID))
			}
		}
	}
	res.Users, _ = repo.Users.GetMany(mUsers.ToArray())
	if len(res.Users) != len(mUsers) {
		logs.Warn("River found unmatched dialog users", zap.Int("Got", len(res.Users)), zap.Int("Need", len(mUsers)))
		for userID := range mUsers {
			found := false
			for _, g := range res.Users {
				if g.ID == userID {
					found = true
					break
				}
			}
			if !found {
				logs.Warn("missed user", zap.Int64("UserID", userID))
			}
		}
	}

	out.Constructor = msg.C_MessagesDialogs
	buff, err := res.Marshal()
	logs.ErrorOnErr("River got error on marshal MessagesDialogs", err)
	out.Message = buff
	da.OnComplete(out)
}
