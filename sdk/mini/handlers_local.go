package mini

import (
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/uiexec"
	"github.com/ronaksoft/rony"
	"go.uber.org/zap"
	"sort"
)

/*
   Creation Time: 2021 - May - 01
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func (r *River) accountsGetTeams(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.AccountGetTeams{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	teams := repo.Teams.List()

	if len(teams) > 0 {
		teamsMany := &msg.TeamsMany{
			Teams: teams,
		}
		out.Fill(out.RequestID, msg.C_TeamsMany, teamsMany)
		successCB(out)
		return
	}

	r.networkCtrl.HttpCommand(in, timeoutCB, successCB)
}

func (r *River) messagesSendMedia(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesSendMedia{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	switch req.MediaType {
	case msg.InputMediaType_InputMediaTypeContact, msg.InputMediaType_InputMediaTypeGeoLocation,
		msg.InputMediaType_InputMediaTypeDocument, msg.InputMediaType_InputMediaTypeMessageDocument:
		// This will be used as next requestID
		req.RandomID = domain.SequentialUniqueID()

		// Insert into pending messages, id is negative nano timestamp and save RandomID too : Done
		dbID := -req.RandomID

		res, err := repo.PendingMessages.SaveMessageMedia(domain.GetTeamID(in), domain.GetTeamAccess(in), dbID, r.ConnInfo.UserID, req)
		if err != nil {
			e := &rony.Error{
				Code:  "n/a",
				Items: "Failed to save to pendingMessages : " + err.Error(),
			}
			out.Fill(out.RequestID, rony.C_Error, e)
			uiexec.ExecSuccessCB(successCB, out)
			return
		}
		// Return to CallBack with pending message data : Done
		out.Constructor = msg.C_ClientPendingMessage

		out.Message, _ = res.Marshal()
		uiexec.ExecSuccessCB(successCB, out)

	case msg.InputMediaType_InputMediaTypeUploadedDocument:
		// no need to insert pending message cuz we already insert one b4 start uploading
	}

	requestBytes, _ := req.Marshal()
	r.networkCtrl.HttpCommand(
		&rony.MessageEnvelope{
			Constructor: msg.C_MessagesSendMedia,
			RequestID:   uint64(req.RandomID),
			Message:     requestBytes,
			Header:      in.Header,
		}, timeoutCB, successCB,
	)
}

func (r *River) messagesGetDialogs(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesGetDialogs{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}
	res := &msg.MessagesDialogs{}
	res.Dialogs = repo.Dialogs.List(domain.GetTeamID(in), req.Offset, req.Limit)
	res.Count = repo.Dialogs.CountDialogs(domain.GetTeamID(in))

	pendingMessages := repo.PendingMessages.GetAndConvertAll()
	dialogPMs := make(map[string]*msg.UserMessage, len(pendingMessages))
	for _, pm := range pendingMessages {
		keyID := fmt.Sprintf("%d.%d", pm.PeerID, pm.PeerType)
		v, ok := dialogPMs[keyID]
		if !ok {
			dialogPMs[keyID] = pm
		} else if pm.ID < v.ID {
			dialogPMs[keyID] = pm
		}
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
		keyID := fmt.Sprintf("%d.%d", dialog.PeerID, dialog.PeerType)
		if pm, ok := dialogPMs[keyID]; ok {
			dialog.TopMessageID = pm.ID
		}
	}

	// Load Messages
	res.Messages = repo.Messages.GetMany(mMessages.ToArray())

	// Load Pending messages
	res.Messages = append(res.Messages, pendingMessages...)
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
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) messagesGetDialog(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesGetDialog{}
	err := req.Unmarshal(in.Message)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}
	res := &msg.Dialog{}
	res, err = repo.Dialogs.Get(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))

	// if the localDB had no data send the request to server
	if err != nil {
		logs.Warn("We got error on repo GetDialog", zap.Error(err), zap.Int64("PeerID", req.Peer.ID))
		r.networkCtrl.HttpCommand(in, timeoutCB, successCB)
		return
	}

	out.Constructor = msg.C_Dialog
	out.Message, _ = res.Marshal()

	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) messagesDelete(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesDelete{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	// send the request to server
	r.networkCtrl.HttpCommand(in, timeoutCB, successCB)

}

func (r *River) messagesGet(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesGet{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}
	msgIDs := domain.MInt64B{}
	pMsgIDs := domain.MInt64B{}

	for _, v := range req.MessagesIDs {
		if v > 0 {
			msgIDs[v] = true
		} else {
			pMsgIDs[v] = true
		}
	}

	messages := repo.Messages.GetMany(msgIDs.ToArray())
	messages = append(messages, repo.PendingMessages.GetMany(pMsgIDs.ToArray())...)

	mUsers := domain.MInt64B{}
	mUsers[req.Peer.ID] = true
	for _, m := range messages {
		mUsers[m.SenderID] = true
		mUsers[m.FwdSenderID] = true
		actUserIDs := domain.ExtractActionUserIDs(m.MessageAction, m.MessageActionData)
		for _, id := range actUserIDs {
			mUsers[id] = true
		}
	}
	users, _ := repo.Users.GetMany(mUsers.ToArray())

	// if db already had all users
	if len(messages) == (len(msgIDs)+len(pMsgIDs)) && len(users) > 0 {
		res := &msg.MessagesMany{
			Messages: messages,
			Users:    users,
		}

		out.Constructor = msg.C_MessagesMany
		out.Message, _ = res.Marshal()
		uiexec.ExecSuccessCB(successCB, out)
		return
	}

	// send the request to the server
	r.networkCtrl.HttpCommand(in, timeoutCB, successCB)
}

func (r *River) usersGetFull(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.UsersGetFull{}
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::usersGetFull()-> Unmarshal()", zap.Error(err))
		return
	}
	userIDs := domain.MInt64B{}
	for _, v := range req.Users {
		userIDs[v.UserID] = true
	}

	users, _, _ := repo.Users.GetManyWithOutdated(userIDs.ToArray())
	allResolved := len(users) == len(userIDs)
	if allResolved {
		res := &msg.UsersMany{}
		for _, user := range users {
			user.PhotoGallery = repo.Users.GetPhotoGallery(user.ID)
			sort.Slice(user.PhotoGallery, func(i, j int) bool {
				return user.PhotoGallery[i].PhotoID > user.PhotoGallery[j].PhotoID
			})
		}
		res.Users = users

		out.Constructor = msg.C_UsersMany
		out.Message, _ = res.Marshal()
		uiexec.ExecSuccessCB(successCB, out)

		return
	}

	// send the request to server
	r.networkCtrl.HttpCommand(in, timeoutCB, successCB)
}

func (r *River) usersGet(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.UsersGet{}
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::usersGet()-> Unmarshal()", zap.Error(err))
		return
	}
	userIDs := domain.MInt64B{}
	for _, v := range req.Users {
		userIDs[v.UserID] = true
	}

	users, _, _ := repo.Users.GetManyWithOutdated(userIDs.ToArray())
	allResolved := len(users) == len(userIDs)
	if allResolved {
		res := new(msg.UsersMany)
		res.Users = users

		out.Constructor = msg.C_UsersMany
		out.Message, _ = res.Marshal()
		uiexec.ExecSuccessCB(successCB, out)

		return
	}

	// send the request to server
	r.networkCtrl.HttpCommand(in, timeoutCB, successCB)
}
