package search

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/uiexec"
	"github.com/ronaksoft/rony"
	"strings"
	"time"
)

/*
   Creation Time: 2021 - May - 10
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func (r *search) clientGlobalSearch(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ClientGlobalSearch{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	searchPhrase := strings.ToLower(req.Text)
	searchResults := &msg.ClientSearchResult{}
	var userContacts []*msg.ContactUser
	var nonContacts []*msg.ContactUser
	var msgs []*msg.UserMessage
	if len(req.LabelIDs) > 0 {
		if req.Peer != nil {
			msgs = repo.Messages.SearchByLabels(domain.GetTeamID(in), req.LabelIDs, req.Peer.ID, req.Limit)
		} else {
			msgs = repo.Messages.SearchByLabels(domain.GetTeamID(in), req.LabelIDs, 0, req.Limit)
		}

	} else if req.SenderID != 0 {
		msgs = repo.Messages.SearchBySender(domain.GetTeamID(in), searchPhrase, req.SenderID, req.Peer.ID, req.Limit)
	} else if req.Peer != nil {
		msgs = repo.Messages.SearchTextByPeerID(domain.GetTeamID(in), searchPhrase, req.Peer.ID, req.Limit)
	} else {
		msgs = repo.Messages.SearchText(domain.GetTeamID(in), searchPhrase, req.Limit)
	}

	// get users && group IDs
	userIDs := domain.MInt64B{}
	matchedUserIDs := domain.MInt64B{}
	groupIDs := domain.MInt64B{}
	for _, m := range msgs {
		if m.PeerType == int32(msg.PeerType_PeerSelf) || m.PeerType == int32(msg.PeerType_PeerUser) {
			userIDs[m.PeerID] = true
		}
		if m.PeerType == int32(msg.PeerType_PeerGroup) {
			groupIDs[m.PeerID] = true
		}
		if m.SenderID > 0 {
			userIDs[m.SenderID] = true
		} else {
			groupIDs[m.PeerID] = true
		}
		if m.FwdSenderID > 0 {
			userIDs[m.FwdSenderID] = true
		} else {
			groupIDs[m.FwdSenderID] = true
		}
	}

	// if peerID == 0 then look for group and contact names too
	if req.Peer == nil {
		userContacts, _ = repo.Users.SearchContacts(domain.GetTeamID(in), searchPhrase)
		for _, userContact := range userContacts {
			matchedUserIDs[userContact.ID] = true
		}
		nonContacts = repo.Users.SearchNonContacts(domain.GetTeamID(in), searchPhrase)
		for _, userContact := range nonContacts {
			matchedUserIDs[userContact.ID] = true
		}
		searchResults.MatchedGroups = repo.Groups.Search(domain.GetTeamID(in), searchPhrase)
	}

	users, _ := repo.Users.GetMany(userIDs.ToArray())
	groups, _ := repo.Groups.GetMany(groupIDs.ToArray())
	matchedUsers, _ := repo.Users.GetMany(matchedUserIDs.ToArray())

	searchResults.Messages = msgs
	searchResults.Users = users
	searchResults.Groups = groups
	searchResults.MatchedUsers = matchedUsers
	out.Fill(in.RequestID, msg.C_ClientSearchResult, searchResults)
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *search) clientGetRecentSearch(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ClientGetRecentSearch{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	recentSearches := repo.RecentSearches.List(domain.GetTeamID(in), req.Limit)

	// get users && group IDs
	userIDs := domain.MInt64B{}
	groupIDs := domain.MInt64B{}
	for _, r := range recentSearches {
		if r.Peer.Type == int32(msg.PeerType_PeerSelf) || r.Peer.Type == int32(msg.PeerType_PeerUser) {
			userIDs[r.Peer.ID] = true
		}
		if r.Peer.Type == int32(msg.PeerType_PeerGroup) {
			groupIDs[r.Peer.ID] = true
		}
	}

	users, _ := repo.Users.GetMany(userIDs.ToArray())
	groups, _ := repo.Groups.GetMany(groupIDs.ToArray())

	res := &msg.ClientRecentSearchMany{
		RecentSearches: recentSearches,
		Users:          users,
		Groups:         groups,
	}
	out.Fill(in.RequestID, msg.C_ClientRecentSearchMany, res)
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *search) clientPutRecentSearch(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ClientPutRecentSearch{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	peer := &msg.Peer{
		ID:         req.Peer.ID,
		Type:       int32(req.Peer.Type),
		AccessHash: req.Peer.AccessHash,
	}

	recentSearch := &msg.ClientRecentSearch{
		Peer: peer,
		Date: int32(time.Now().Unix()),
	}

	err := repo.RecentSearches.Put(domain.GetTeamID(in), recentSearch)

	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	res := &msg.Bool{
		Result: true,
	}
	out.Fill(in.RequestID, msg.C_Bool, res)
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *search) clientRemoveAllRecentSearches(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ClientRemoveAllRecentSearches{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	err := repo.RecentSearches.Clear(domain.GetTeamID(in))

	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	res := &msg.Bool{
		Result: true,
	}
	out.Fill(in.RequestID, msg.C_Bool, res)
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *search) clientRemoveRecentSearch(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ClientRemoveRecentSearch{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	err := repo.RecentSearches.Delete(domain.GetTeamID(in), req.Peer)

	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	res := &msg.Bool{
		Result: true,
	}
	out.Fill(in.RequestID, msg.C_Bool, res)
	uiexec.ExecSuccessCB(successCB, out)
}
