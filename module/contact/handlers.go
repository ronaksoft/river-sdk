package contact

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/request"
	"git.ronaksoft.com/river/sdk/internal/uiexec"
	"github.com/ronaksoft/rony"
	"go.uber.org/zap"
	"strings"
)

/*
   Creation Time: 2021 - May - 10
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func (r *contact) contactsGet(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.ContactsGet{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	res := &msg.ContactsMany{}
	res.ContactUsers, res.Contacts = repo.Users.GetContacts(domain.GetTeamID(in))

	userIDs := make([]int64, 0, len(res.ContactUsers))
	for idx := range res.ContactUsers {
		userIDs = append(userIDs, res.ContactUsers[idx].ID)
	}
	res.Users, _ = repo.Users.GetMany(userIDs)
	out.Constructor = msg.C_ContactsMany
	out.Message, _ = res.Marshal()

	r.Log().Info("returned data locally, ContactsGet",
		zap.Int("Users", len(res.Users)),
		zap.Int("Contacts", len(res.Contacts)),
	)
	uiexec.ExecSuccessCB(da.OnComplete, out)
}

func (r *contact) contactsAdd(in, out *rony.MessageEnvelope, da request.Callback) {
	if domain.GetTeamID(in) != 0 {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: "teams cannot add contact"})
		da.OnComplete(out)
		return
	}

	req := &msg.ContactsAdd{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	user, _ := repo.Users.Get(req.User.UserID)
	if user != nil {
		user.FirstName = req.FirstName
		user.LastName = req.LastName
		user.Phone = req.Phone
		_ = repo.Users.SaveContact(domain.GetTeamID(in), &msg.ContactUser{
			ID:         user.ID,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			AccessHash: user.AccessHash,
			Phone:      user.Phone,
			Username:   user.Username,
			ClientID:   0,
			Photo:      user.Photo,
		})
		_ = repo.Users.Save(user)
	}

	// reset contacts hash to update the contacts
	_ = repo.System.SaveInt(domain.GetContactsGetHashKey(domain.GetTeamID(in)), 0)
	r.SDK().QueueCtrl().EnqueueCommand(in, da.OnTimeout, da.OnComplete, da.UI())
}

func (r *contact) contactsImport(in, out *rony.MessageEnvelope, da request.Callback) {
	if domain.GetTeamID(in) != 0 {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: "teams cannot import contact"})
		da.OnComplete(out)
		return
	}

	req := &msg.ContactsImport{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	// If only importing one contact then we don't need to calculate contacts hash
	if len(req.Contacts) == 1 {
		// send request to server
		r.SDK().QueueCtrl().EnqueueCommand(in, da.OnTimeout, da.OnComplete, da.UI())
		return
	}

	oldHash, err := repo.System.LoadInt(domain.SkContactsImportHash)
	if err != nil {
		r.Log().Warn("got error on loading ContactsImportHash", zap.Error(err))
	}
	// calculate ContactsImportHash and compare with oldHash
	newHash := domain.CalculateContactsImportHash(req)
	r.Log().Info("returned data locally, ContactsImport",
		zap.Uint64("Old", oldHash),
		zap.Uint64("New", newHash),
	)
	if newHash == oldHash {
		res := &msg.ContactsImported{
			ContactUsers: nil,
			Users:        nil,
			Empty:        true,
		}
		out.Fill(out.RequestID, msg.C_ContactsImported, res)
		da.OnComplete(out)
		return
	}

	// not equal save it to DB
	err = repo.System.SaveInt(domain.SkContactsImportHash, newHash)
	if err != nil {
		r.Log().Error("got error on saving ContactsImportHash", zap.Error(err))
	}

	// extract differences between existing contacts and new contacts
	_, contacts := repo.Users.GetContacts(domain.GetTeamID(in))
	diffContacts := domain.ExtractsContactsDifference(contacts, req.Contacts)

	err = repo.Users.SavePhoneContact(diffContacts...)
	if err != nil {
		r.Log().Error("got error on saving phone contacts in to the db", zap.Error(err))
	}

	if len(diffContacts) <= 250 {
		// send the request to server
		r.SDK().QueueCtrl().EnqueueCommand(in, da.OnTimeout, da.OnComplete, da.UI())
		return
	}

	// chunk contacts by size of 50 and send them to server
	r.SDK().SyncCtrl().ContactsImport(req.Replace, da.OnComplete, out)
}

func (r *contact) contactsDelete(in, out *rony.MessageEnvelope, da request.Callback) {
	if domain.GetTeamID(in) != 0 {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: "teams cannot delete contact"})
		da.OnComplete(out)
		return
	}

	req := &msg.ContactsDelete{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	_ = repo.Users.DeleteContact(domain.GetTeamID(in), req.UserIDs...)
	_ = repo.System.SaveInt(domain.GetContactsGetHashKey(domain.GetTeamID(in)), 0)

	r.SDK().QueueCtrl().EnqueueCommand(in, da.OnTimeout, da.OnComplete, da.UI())
	return
}

func (r *contact) contactsDeleteAll(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.ContactsDeleteAll{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	_ = repo.Users.DeleteAllContacts(domain.GetTeamID(in))
	_ = repo.System.SaveInt(domain.GetContactsGetHashKey(domain.GetTeamID(in)), 0)
	_ = repo.System.SaveInt(domain.SkContactsImportHash, 0)
	r.SDK().QueueCtrl().EnqueueCommand(in, da.OnTimeout, da.OnComplete, da.UI())
	return
}

func (r *contact) contactsGetTopPeers(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.ContactsGetTopPeers{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}
	res := &msg.ContactsTopPeers{}
	topPeers, _ := repo.TopPeers.List(domain.GetTeamID(in), req.Category, req.Offset, req.Limit)
	if len(topPeers) == 0 {
		r.SDK().QueueCtrl().EnqueueCommand(in, da.OnTimeout, da.OnComplete, da.UI())
		return
	}

	res.Category = req.Category
	res.Peers = topPeers
	res.Count = int32(len(topPeers))

	mUsers := domain.MInt64B{}
	mGroups := domain.MInt64B{}
	for _, topPeer := range res.Peers {
		switch msg.PeerType(topPeer.Peer.Type) {
		case msg.PeerType_PeerUser, msg.PeerType_PeerExternalUser:
			mUsers[topPeer.Peer.ID] = true
		case msg.PeerType_PeerGroup:
			mGroups[topPeer.Peer.ID] = true
		}
	}
	res.Groups, _ = repo.Groups.GetMany(mGroups.ToArray())
	if len(res.Groups) != len(mGroups) {
		r.Log().Warn("found unmatched top peers groups", zap.Int("Got", len(res.Groups)), zap.Int("Need", len(mGroups)))
		for groupID := range mGroups {
			found := false
			for _, g := range res.Groups {
				if g.ID == groupID {
					found = true
					break
				}
			}
			if !found {
				r.Log().Warn("missed group", zap.Int64("GroupID", groupID))
			}
		}
	}

	res.Users, _ = repo.Users.GetMany(mUsers.ToArray())
	if len(res.Users) != len(mUsers) {
		r.Log().Warn("found unmatched top peers users", zap.Int("Got", len(res.Users)), zap.Int("Need", len(mUsers)))
		for userID := range mUsers {
			found := false
			for _, g := range res.Users {
				if g.ID == userID {
					found = true
					break
				}
			}
			if !found {
				r.Log().Warn("missed user", zap.Int64("UserID", userID))
			}
		}
	}

	out.Constructor = msg.C_ContactsTopPeers
	buff, err := res.Marshal()
	r.Log().ErrorOnErr("got error on marshal ContactsTopPeers", err)
	out.Message = buff
	uiexec.ExecSuccessCB(da.OnComplete, out)
}

func (r *contact) contactsResetTopPeer(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.ContactsResetTopPeer{}
	err := req.Unmarshal(in.Message)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	err = repo.TopPeers.Delete(req.Category, domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	r.SDK().QueueCtrl().EnqueueCommand(in, da.OnTimeout, da.OnComplete, da.UI())
}

func (r *contact) clientContactSearch(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.ClientContactSearch{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	searchPhrase := strings.ToLower(req.Text)
	r.Log().Info("SearchContacts", zap.String("Phrase", searchPhrase))

	users := &msg.UsersMany{}
	contactUsers, _ := repo.Users.SearchContacts(domain.GetTeamID(in), searchPhrase)
	userIDs := make([]int64, 0, len(contactUsers))
	for _, contactUser := range contactUsers {
		userIDs = append(userIDs, contactUser.ID)
	}
	users.Users, _ = repo.Users.GetMany(userIDs)

	out.Fill(in.RequestID, msg.C_UsersMany, users)
	uiexec.ExecSuccessCB(da.OnComplete, out)
}
