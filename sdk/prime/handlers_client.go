package riversdk

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
	"strings"
	"time"
)

/*
   Creation Time: 2020 - Nov - 05
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func (r *River) clientGetMediaHistory(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ClientGetMediaHistory{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	messages, users, groups := repo.Messages.GetMediaMessageHistory(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit, req.Cat)
	if len(messages) > 0 {
		res := &msg.MessagesMany{
			Messages: messages,
			Users:    users,
			Groups:   groups,
		}

		out.RequestID = in.RequestID
		out.Constructor = msg.C_MessagesMany
		out.Message, _ = res.Marshal()
		uiexec.ExecSuccessCB(successCB, out)
		return
	}
}

func (r *River) clientSendMessageMedia(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	reqMedia := &msg.ClientSendMessageMedia{}
	if err := reqMedia.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		uiexec.ExecSuccessCB(successCB, out)
		return
	}

	// support IOS file path
	if strings.HasPrefix(reqMedia.FilePath, "file://") {
		reqMedia.FilePath = reqMedia.FilePath[7:]
	}
	if strings.HasPrefix(reqMedia.ThumbFilePath, "file://") {
		reqMedia.ThumbFilePath = reqMedia.ThumbFilePath[7:]
	}

	// 1. insert into pending messages, id is negative nano timestamp and save RandomID too : Done
	fileID := domain.SequentialUniqueID()
	msgID := -fileID
	thumbID := int64(0)
	reqMedia.FileUploadID = fmt.Sprintf("%d", fileID)
	reqMedia.FileID = fileID
	if reqMedia.ThumbFilePath != "" {
		thumbID = domain.RandomInt63()
		reqMedia.ThumbID = thumbID
		reqMedia.ThumbUploadID = fmt.Sprintf("%d", thumbID)
	}

	checkSha256 := true
	switch reqMedia.MediaType {
	case msg.InputMediaType_InputMediaTypeUploadedDocument:
		for _, attr := range reqMedia.Attributes {
			if attr.Type == msg.DocumentAttributeType_AttributeTypeAudio {
				x := &msg.DocumentAttributeAudio{}
				_ = x.Unmarshal(attr.Data)
				if x.Voice {
					checkSha256 = false
				}
			}
		}
	default:
		panic("Invalid MediaInputType")
	}

	h, _ := domain.CalculateSha256(reqMedia.FilePath)
	pendingMessage, err := repo.PendingMessages.SaveClientMessageMedia(
		domain.GetTeamID(in), domain.GetTeamAccess(in), msgID, r.ConnInfo.UserID, fileID, fileID, thumbID, reqMedia, h,
	)
	if err != nil {
		e := &rony.Error{
			Code:  "n/a",
			Items: "Failed to save to pendingMessages : " + err.Error(),
		}
		out.Fill(out.RequestID, rony.C_Error, e)
		uiexec.ExecSuccessCB(successCB, out)
		return
	}

	// 3. return to CallBack with pending message data : Done
	out.Fill(out.RequestID, msg.C_ClientPendingMessage, pendingMessage)

	// 4. Start the upload process
	r.fileCtrl.UploadMessageDocument(pendingMessage.ID, reqMedia.FilePath, reqMedia.ThumbFilePath, fileID, thumbID, h, pendingMessage.PeerID, checkSha256)

	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) clientGlobalSearch(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
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
		nonContacts = repo.Users.SearchNonContacts(searchPhrase)
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

func (r *River) clientContactSearch(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ClientContactSearch{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	searchPhrase := strings.ToLower(req.Text)
	logs.Info("SearchContacts", zap.String("Phrase", searchPhrase))

	users := &msg.UsersMany{}
	contactUsers, _ := repo.Users.SearchContacts(domain.GetTeamID(in), searchPhrase)
	userIDs := make([]int64, 0, len(contactUsers))
	for _, contactUser := range contactUsers {
		userIDs = append(userIDs, contactUser.ID)
	}
	users.Users, _ = repo.Users.GetMany(userIDs)

	out.Fill(in.RequestID, msg.C_UsersMany, users)
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) clientGetCachedMedia(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ClientGetCachedMedia{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	res := repo.Files.GetCachedMedia(domain.GetTeamID(in))

	out.Fill(in.RequestID, msg.C_ClientCachedMediaInfo, res)
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) clientClearCachedMedia(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ClientClearCachedMedia{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	if req.Peer != nil {
		repo.Files.DeleteCachedMediaByPeer(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.MediaTypes)
	} else if len(req.MediaTypes) > 0 {
		repo.Files.DeleteCachedMediaByMediaType(domain.GetTeamID(in), req.MediaTypes)
	} else {
		repo.Files.ClearCache()
	}

	res := &msg.Bool{
		Result: true,
	}
	out.Fill(in.RequestID, msg.C_Bool, res)
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) clientGetAllDownloadedMedia(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ClientGetAllDownloadedMedia{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	msgs, _ := repo.Messages.GetAllMedia(req.MediaType)

	// get users && group IDs
	userIDs := domain.MInt64B{}
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
		}
		if m.FwdSenderID > 0 {
			userIDs[m.FwdSenderID] = true
		}
	}

	users, _ := repo.Users.GetMany(userIDs.ToArray())
	groups, _ := repo.Groups.GetMany(groupIDs.ToArray())

	res := &msg.MessagesMany{
		Messages:   msgs,
		Users:      users,
		Groups:     groups,
		Continuous: false,
	}
	out.Fill(in.RequestID, msg.C_MessagesMany, res)
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) clientGetLastBotKeyboard(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ClientGetLastBotKeyboard{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	lastKeyboardMsg, _ := repo.Messages.GetLastBotKeyboard(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))

	if lastKeyboardMsg == nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: "message not found"})
		successCB(out)
		return
	}

	out.Fill(in.RequestID, msg.C_UserMessage, lastKeyboardMsg)
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) clientGetRecentSearch(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
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

func (r *River) clientPutRecentSearch(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
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

func (r *River) clientRemoveAllRecentSearches(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
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

func (r *River) clientRemoveRecentSearch(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
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

func (r *River) clientGetTeamCounters(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ClientGetTeamCounters{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	unreadCount, mentionCount, err := repo.Dialogs.CountAllUnread(r.ConnInfo.UserID, req.Team.ID, req.WithMutes)

	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	res := &msg.ClientTeamCounters{
		UnreadCount:  int64(unreadCount),
		MentionCount: int64(mentionCount),
	}

	out.Fill(in.RequestID, msg.C_ClientTeamCounters, res)
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) clientGetFrequentReactions(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	reactions := domain.SysConfig.Reactions
	logs.Info("Reactions", zap.Int("ReactionsCount", len(reactions)))

	useCountsMap := make(map[string]uint32, len(reactions))

	for _, r := range reactions {
		useCount, _ := repo.Reactions.GetReactionUseCount(r)
		useCountsMap[r] = useCount
	}

	sort.Slice(reactions, func(i, j int) bool {
		return useCountsMap[reactions[i]] > useCountsMap[reactions[j]]
	})

	res := &msg.ClientFrequentReactions{
		Reactions: reactions,
	}
	out.Fill(out.RequestID, msg.C_ClientFrequentReactions, res)
	successCB(out)
}
