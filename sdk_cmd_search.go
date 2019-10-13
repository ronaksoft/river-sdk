package riversdk

import (
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	mon "git.ronaksoftware.com/ronak/riversdk/pkg/monitoring"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"go.uber.org/zap"
	"strings"
	"time"
)

/*
   Creation Time: 2019 - Jul - 24
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

// SearchContacts searches contacts
func (r *River) SearchContacts(requestID int64, searchPhrase string, delegate RequestDelegate) {
	searchPhrase = strings.ToLower(searchPhrase)
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("SearchContacts", time.Now().Sub(startTime))
	}()
	logs.Info("SearchContacts", zap.String("Phrase", searchPhrase))
	res := new(msg.MessageEnvelope)
	res.Constructor = msg.C_UsersMany
	res.RequestID = uint64(requestID)
	users := new(msg.UsersMany)

	contactUsers, _ := repo.Users.SearchContacts(searchPhrase)
	userIDs := make([]int64, 0, len(contactUsers))
	for _, contactUser := range contactUsers {
		userIDs = append(userIDs, contactUser.ID)
	}
	users.Users = repo.Users.GetMany(userIDs)
	res.Message, _ = users.Marshal()
	buff, _ := res.Marshal()
	if delegate != nil {
		delegate.OnComplete(buff)
	}
}

// SearchGlobal returns messages, contacts and groups matching given text
// peerID 0 means search is not limited to a specific peerID
func (r *River) SearchGlobal(searchPhrase string, peerID int64, delegate RequestDelegate) {
	searchPhrase = strings.ToLower(searchPhrase)
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("SearchGlobal", time.Now().Sub(startTime))
	}()
	searchResults := new(msg.ClientSearchResult)
	var userContacts []*msg.ContactUser
	var nonContacts []*msg.ContactUser
	var msgs []*msg.UserMessage
	if peerID != 0 {
		msgs = repo.Messages.SearchTextByPeerID(searchPhrase, peerID)
	} else {
		msgs = repo.Messages.SearchText(searchPhrase)
	}

	// get users && group IDs
	userIDs := domain.MInt64B{}
	matchedUserIDs := domain.MInt64B{}
	groupIDs := domain.MInt64B{}
	for _, m := range msgs {
		if m.PeerType == int32(msg.PeerSelf) || m.PeerType == int32(msg.PeerUser) {
			userIDs[m.PeerID] = true
		}
		if m.PeerType == int32(msg.PeerGroup) {
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
	if peerID == 0 {
		userContacts, _ = repo.Users.SearchContacts(searchPhrase)
		for _, userContact := range userContacts {
			matchedUserIDs[userContact.ID] = true
		}
		nonContacts = repo.Users.SearchNonContacts(searchPhrase)
		for _, userContact := range nonContacts {
			matchedUserIDs[userContact.ID] = true
		}
		searchResults.MatchedGroups = repo.Groups.Search(searchPhrase)
	}

	users := repo.Users.GetMany(userIDs.ToArray())
	groups := repo.Groups.GetMany(groupIDs.ToArray())
	matchedUsers := repo.Users.GetMany(matchedUserIDs.ToArray())

	searchResults.Messages = msgs
	searchResults.Users = users
	searchResults.Groups = groups
	searchResults.MatchedUsers = matchedUsers

	outBytes, _ := searchResults.Marshal()
	if delegate != nil {
		delegate.OnComplete(outBytes)
	}
}


func (r *River) SearchReIndex() {
	repo.Users.ReIndex()
	repo.Groups.ReIndex()
	repo.Messages.ReIndex()
}