package contact

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/request"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/errors"
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

func (r *contact) contactsGet(da request.Callback) {
	req := &msg.ContactsGet{}
	if err := da.RequestData(req); err != nil {
		return
	}

	res := &msg.ContactsMany{}
	res.ContactUsers, res.Contacts = repo.Users.GetContacts(da.TeamID())

	userIDs := make([]int64, 0, len(res.ContactUsers))
	for idx := range res.ContactUsers {
		userIDs = append(userIDs, res.ContactUsers[idx].ID)
	}
	res.Users, _ = repo.Users.GetMany(userIDs)

	r.Log().Info("returned data locally, ContactsGet",
		zap.Int("Users", len(res.Users)),
		zap.Int("Contacts", len(res.Contacts)),
	)

	da.Response(msg.C_ContactsMany, res)
}

func (r *contact) contactsAdd(da request.Callback) {
	req := &msg.ContactsAdd{}
	if err := da.RequestData(req); err != nil {
		return
	}

	if da.TeamID() != 0 {
		out := &rony.MessageEnvelope{}
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: "teams cannot add contact"}, da.Envelope().Header...)
		da.OnComplete(out)
		return
	}

	user, _ := repo.Users.Get(req.User.UserID)
	if user != nil {
		user.FirstName = req.FirstName
		user.LastName = req.LastName
		user.Phone = req.Phone
		_ = repo.Users.SaveContact(da.TeamID(), &msg.ContactUser{
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
	_ = repo.System.SaveInt(domain.GetContactsGetHashKey(da.TeamID()), 0)
	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *contact) contactsImport(da request.Callback) {
	req := &msg.ContactsImport{}
	if err := da.RequestData(req); err != nil {
		return
	}

	if da.TeamID() != 0 {
		da.OnComplete(errors.Message(da.RequestID(), "00", "teams cannot import contact"))
		return
	}

	// If only importing one contact then we don't need to calculate contacts hash
	if len(req.Contacts) == 1 {
		// send request to server
		r.SDK().QueueCtrl().EnqueueCommand(da)
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
		da.Response(msg.C_ContactsImported, res)
		return
	}

	// not equal save it to DB
	err = repo.System.SaveInt(domain.SkContactsImportHash, newHash)
	if err != nil {
		r.Log().Error("got error on saving ContactsImportHash", zap.Error(err))
	}

	// extract differences between existing contacts and new contacts
	_, contacts := repo.Users.GetContacts(da.TeamID())
	diffContacts := domain.ExtractsContactsDifference(contacts, req.Contacts)

	err = repo.Users.SavePhoneContact(diffContacts...)
	if err != nil {
		r.Log().Error("got error on saving phone contacts in to the db", zap.Error(err))
	}

	if len(diffContacts) <= 250 {
		// send the request to server
		r.SDK().QueueCtrl().EnqueueCommand(da)
		return
	}

	// chunk contacts by size of 50 and send them to server
	r.SDK().SyncCtrl().ContactsImport(req.Replace, da.OnComplete)
}

func (r *contact) contactsDelete(da request.Callback) {
	if da.TeamID() != 0 {
		da.OnComplete(errors.Message(da.RequestID(), "00", "teams cannot delete contact"))
		return
	}

	req := &msg.ContactsDelete{}
	if err := da.RequestData(req); err != nil {
		return
	}

	_ = repo.Users.DeleteContact(da.TeamID(), req.UserIDs...)
	_ = repo.System.SaveInt(domain.GetContactsGetHashKey(da.TeamID()), 0)

	r.SDK().QueueCtrl().EnqueueCommand(da)
	return
}

func (r *contact) contactsDeleteAll(da request.Callback) {
	req := &msg.ContactsDeleteAll{}
	if err := da.RequestData(req); err != nil {
		return
	}

	_ = repo.Users.DeleteAllContacts(da.TeamID())
	_ = repo.System.SaveInt(domain.GetContactsGetHashKey(da.TeamID()), 0)
	_ = repo.System.SaveInt(domain.SkContactsImportHash, 0)
	r.SDK().QueueCtrl().EnqueueCommand(da)
	return
}

func (r *contact) contactsGetTopPeers(da request.Callback) {
	req := &msg.ContactsGetTopPeers{}
	if err := da.RequestData(req); err != nil {
		return
	}

	res := &msg.ContactsTopPeers{}
	topPeers, _ := repo.TopPeers.List(da.TeamID(), req.Category, req.Offset, req.Limit)
	if len(topPeers) == 0 {
		r.SDK().QueueCtrl().EnqueueCommand(da)
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

	da.Response(msg.C_ContactsTopPeers, res)
}

func (r *contact) contactsResetTopPeer(da request.Callback) {
	req := &msg.ContactsResetTopPeer{}
	if err := da.RequestData(req); err != nil {
		return
	}

	err := repo.TopPeers.Delete(req.Category, da.TeamID(), req.Peer.ID, int32(req.Peer.Type))
	if err != nil {
		da.OnComplete(errors.Message(da.RequestID(), "00", err.Error()))
		return
	}

	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *contact) clientContactSearch(da request.Callback) {
	req := &msg.ClientContactSearch{}
	if err := da.RequestData(req); err != nil {
		return
	}

	searchPhrase := strings.ToLower(req.Text)
	r.Log().Info("SearchContacts", zap.String("Phrase", searchPhrase))

	users := &msg.UsersMany{}
	contactUsers, _ := repo.Users.SearchContacts(da.TeamID(), searchPhrase)
	userIDs := make([]int64, 0, len(contactUsers))
	for _, contactUser := range contactUsers {
		userIDs = append(userIDs, contactUser.ID)
	}
	users.Users, _ = repo.Users.GetMany(userIDs)

	da.Response(msg.C_UsersMany, users)
}
