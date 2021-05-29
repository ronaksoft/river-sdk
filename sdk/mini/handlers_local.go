package mini

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/minirepo"
	"git.ronaksoft.com/river/sdk/internal/request"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/errors"
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

func (r *River) messagesSendMedia(da request.Callback) {
	req := &msg.MessagesSendMedia{}
	if err := da.RequestData(req); err != nil {
		return
	}

	req.RandomID = domain.SequentialUniqueID()
	da.Envelope().Fill(da.RequestID(), msg.C_MessagesSendMedia, req)
	r.network.HttpCommand(nil, da)
}

func (r *River) messagesGetDialogs(da request.Callback) {
	req := &msg.MessagesGetDialogs{}
	if err := da.RequestData(req); err != nil {
		return
	}

	res := &msg.MessagesDialogs{}
	res.Dialogs, _ = minirepo.Dialogs.List(da.TeamID(), req.Offset, req.Limit)

	// If the localDB had no data send the request to server
	if len(res.Dialogs) == 0 {
		r.network.HttpCommand(nil, da)
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

	res.Groups, _ = minirepo.Groups.ReadMany(da.TeamID(), mGroups.ToArray()...)
	if len(res.Groups) != len(mGroups) {
		logger.Warn("found unmatched dialog groups", zap.Int("Got", len(res.Groups)), zap.Int("Need", len(mGroups)))
		for groupID := range mGroups {
			found := false
			for _, g := range res.Groups {
				if g.ID == groupID {
					found = true
					break
				}
			}
			if !found {
				logger.Warn("missed group", zap.Int64("GroupID", groupID))
			}
		}
	}
	res.Users, _ = minirepo.Users.ReadMany(mUsers.ToArray()...)
	if len(res.Users) != len(mUsers) {
		logger.Warn("found unmatched dialog users", zap.Int("Got", len(res.Users)), zap.Int("Need", len(mUsers)))
		for userID := range mUsers {
			found := false
			for _, g := range res.Users {
				if g.ID == userID {
					found = true
					break
				}
			}
			if !found {
				logger.Warn("missed user", zap.Int64("UserID", userID))
			}
		}
	}

	da.Response(msg.C_MessagesDialogs, res)
}

func (r *River) contactsGet(da request.Callback) {
	req := &msg.ContactsGet{}
	if err := da.RequestData(req); err != nil {
		return
	}

	res, err := minirepo.Users.ReadAllContacts(da.TeamID())
	if err != nil {
		da.Response(rony.C_Error, errors.New("00", err.Error()))
		return
	}

	da.Response(msg.C_ContactsMany, res)
}

func (r *River) accountGetTeams(da request.Callback) {
	req := &msg.AccountGetTeams{}
	if err := da.RequestData(req); err != nil {
		return
	}

	teams, err := minirepo.Teams.List()
	if err != nil {
		da.Response(rony.C_Error, errors.New("00", err.Error()))
		return
	}

	res := &msg.TeamsMany{
		Teams: teams,
	}
	da.Response(msg.C_TeamsMany, res)
}
