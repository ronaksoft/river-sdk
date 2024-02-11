package search

import (
    "strings"
    "time"

    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/river-sdk/internal/repo"
    "github.com/ronaksoft/river-sdk/internal/request"
    "github.com/ronaksoft/rony"
    "github.com/ronaksoft/rony/errors"
)

/*
   Creation Time: 2021 - May - 10
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func (r *search) clientGlobalSearch(da request.Callback) {
    req := &msg.ClientGlobalSearch{}
    if err := da.RequestData(req); err != nil {
        return
    }

    searchPhrase := strings.ToLower(req.Text)
    searchResults := &msg.ClientSearchResult{}
    var userContacts []*msg.ContactUser
    var nonContacts []*msg.ContactUser
    var msgs []*msg.UserMessage
    if len(req.LabelIDs) > 0 {
        if req.Peer != nil {
            msgs = repo.Messages.SearchByLabels(da.TeamID(), req.LabelIDs, req.Peer.ID, req.Limit)
        } else {
            msgs = repo.Messages.SearchByLabels(da.TeamID(), req.LabelIDs, 0, req.Limit)
        }

    } else if req.SenderID != 0 {
        msgs = repo.Messages.SearchBySender(da.TeamID(), searchPhrase, req.SenderID, req.Peer.ID, req.Limit)
    } else if req.Peer != nil {
        msgs = repo.Messages.SearchTextByPeerID(da.TeamID(), searchPhrase, req.Peer.ID, req.Limit)
    } else {
        msgs = repo.Messages.SearchText(da.TeamID(), searchPhrase, req.Limit)
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
        userContacts, _ = repo.Users.SearchContacts(da.TeamID(), searchPhrase)
        for _, userContact := range userContacts {
            matchedUserIDs[userContact.ID] = true
        }
        nonContacts = repo.Users.SearchNonContacts(da.TeamID(), searchPhrase)
        for _, userContact := range nonContacts {
            matchedUserIDs[userContact.ID] = true
        }
        searchResults.MatchedGroups = repo.Groups.Search(da.TeamID(), searchPhrase)
    }

    users, _ := repo.Users.GetMany(userIDs.ToArray())
    groups, _ := repo.Groups.GetMany(groupIDs.ToArray())
    matchedUsers, _ := repo.Users.GetMany(matchedUserIDs.ToArray())

    searchResults.Messages = msgs
    searchResults.Users = users
    searchResults.Groups = groups
    searchResults.MatchedUsers = matchedUsers
    da.Response(msg.C_ClientSearchResult, searchResults)
}

func (r *search) clientGetRecentSearch(da request.Callback) {
    req := &msg.ClientGetRecentSearch{}
    if err := da.RequestData(req); err != nil {
        return
    }

    recentSearches := repo.RecentSearches.List(da.TeamID(), req.Limit)

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
    da.Response(msg.C_ClientRecentSearchMany, res)
}

func (r *search) clientPutRecentSearch(da request.Callback) {
    req := &msg.ClientPutRecentSearch{}
    if err := da.RequestData(req); err != nil {
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

    err := repo.RecentSearches.Put(da.TeamID(), recentSearch)

    if err != nil {
        da.Response(rony.C_Error, errors.New("00", err.Error()))
        return
    }

    res := &msg.Bool{
        Result: true,
    }
    da.Response(msg.C_Bool, res)
}

func (r *search) clientRemoveAllRecentSearches(da request.Callback) {
    req := &msg.ClientRemoveAllRecentSearches{}
    if err := da.RequestData(req); err != nil {
        return
    }

    err := repo.RecentSearches.Clear(da.TeamID())

    if err != nil {
        da.Response(rony.C_Error, errors.New("00", err.Error()))
        return
    }

    res := &msg.Bool{
        Result: true,
    }
    da.Response(msg.C_Bool, res)
}

func (r *search) clientRemoveRecentSearch(da request.Callback) {
    req := &msg.ClientRemoveRecentSearch{}
    if err := da.RequestData(req); err != nil {
        return
    }

    err := repo.RecentSearches.Delete(da.TeamID(), req.Peer)

    if err != nil {
        da.Response(rony.C_Error, errors.New("00", err.Error()))
        return
    }

    res := &msg.Bool{
        Result: true,
    }
    da.Response(msg.C_Bool, res)
}
