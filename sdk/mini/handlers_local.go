package mini

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/minirepo"
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

func (r *River) messagesSendMedia(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.MessagesSendMedia{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	req.RandomID = domain.SequentialUniqueID()
	requestBytes, _ := req.Marshal()
	r.network.HttpCommand(
		&rony.MessageEnvelope{
			Constructor: msg.C_MessagesSendMedia,
			RequestID:   uint64(req.RandomID),
			Message:     requestBytes,
			Header:      in.Header,
		}, da.OnTimeout, da.OnComplete,
	)
}

func (r *River) messagesGetDialogs(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.MessagesGetDialogs{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}
	res := &msg.MessagesDialogs{}
	res.Dialogs, _ = minirepo.Dialogs.List(domain.GetTeamID(in), req.Offset, req.Limit)
	// res.Count = minirepo.Dialogs.CountDialogs(domain.GetTeamID(in))

	// If the localDB had no data send the request to server
	if len(res.Dialogs) == 0 {
		r.network.HttpCommand(in, da.OnTimeout, da.OnComplete)
		return
	}

	mUsers := domain.MInt64B{}
	mGroups := domain.MInt64B{}
	for _, dialog := range res.Dialogs {
		switch msg.PeerType(dialog.PeerType) {
		case msg.PeerType_PeerUser, msg.PeerType_PeerExternalUser:
			if dialog.PeerID != 0 {
				mUsers[dialog.PeerID] = true
			}
		case msg.PeerType_PeerGroup:
			if dialog.PeerID != 0 {
				mGroups[dialog.PeerID] = true
			}
		}
	}

	res.Groups, _ = minirepo.Groups.ReadMany(domain.GetTeamID(in), mGroups.ToArray()...)
	if len(res.Groups) != len(mGroups) {
		r.logger.Warn("River found unmatched dialog groups", zap.Int("Got", len(res.Groups)), zap.Int("Need", len(mGroups)))
		for groupID := range mGroups {
			found := false
			for _, g := range res.Groups {
				if g.ID == groupID {
					found = true
					break
				}
			}
			if !found {
				r.logger.Warn("missed group", zap.Int64("GroupID", groupID))
			}
		}
	}
	res.Users, _ = minirepo.Users.ReadMany(mUsers.ToArray()...)
	if len(res.Users) != len(mUsers) {
		r.logger.Warn("River found unmatched dialog users", zap.Int("Got", len(res.Users)), zap.Int("Need", len(mUsers)))
		for userID := range mUsers {
			found := false
			for _, g := range res.Users {
				if g.ID == userID {
					found = true
					break
				}
			}
			if !found {
				r.logger.Warn("missed user", zap.Int64("UserID", userID))
			}
		}
	}

	out.Fill(in.RequestID, msg.C_MessagesDialogs, res)
	da.OnComplete(out)
}

func (r *River) contactsGet(in, out *rony.MessageEnvelope, da domain.Callback) {
	req := &msg.ContactsGet{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}
	res, err := minirepo.Users.ReadAllContacts()
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}
	out.Fill(in.RequestID, msg.C_ContactsMany, res)
	da.OnComplete(out)
}
