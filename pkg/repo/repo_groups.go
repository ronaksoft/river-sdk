package repo

import (
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"github.com/dgraph-io/badger"
	"strings"
)

const (
	prefixGroups             = "GRP"
	prefixGroupsParticipants = "GRP_P"
)

type repoGroups struct {
	*repository
}

func (r *repoGroups) getGroupKey(groupID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%021d", prefixGroups, groupID))
}

func (r *repoGroups) getGroupParticipantKey(groupID, memberID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%021d.%021d", prefixGroupsParticipants, groupID, memberID))
}

func (r *repoGroups) getPrefix(groupID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%021d.", prefixGroupsParticipants, groupID))
}

func (r *repoGroups) getGroupByKey(groupKey []byte) *msg.Group {
	group := new(msg.Group)
	err := r.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get(groupKey)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return group.Unmarshal(val)
		})
	})
	if err != nil {
		return nil
	}
	return group
}

func (r *repoGroups) readFromDb(groupID int64) *msg.Group {
	group := new(msg.Group)
	err := r.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get(r.getGroupKey(groupID))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return group.Unmarshal(val)
		})
	})
	if err != nil {
		return nil
	}
	return group
}

func (r *repoGroups) readFromCache(groupID int64) *msg.Group {
	group := new(msg.Group)
	keyID := fmt.Sprintf("OBJ.GROUP.{%d}", groupID)

	if jsonGroup, err := lCache.Get(keyID); err != nil || len(jsonGroup) == 0 {
		group := r.readFromDb(groupID)
		if group == nil {
			return nil
		}
		jsonGroup, _ = group.Marshal()
		_ = lCache.Set(keyID, jsonGroup)
		return group
	} else {
		_ = group.Unmarshal(jsonGroup)
	}
	return group
}

func (r *repoGroups) readManyFromCache(groupIDs []int64) []*msg.Group {
	groups := make([]*msg.Group, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		if group := r.readFromCache(groupID); group != nil {
			groups = append(groups, group)
		}
	}
	return groups
}

func (r *repoGroups) deleteFromCache(groupIDs ...int64) {
	for _, groupID := range groupIDs {
		_ = lCache.Delete(fmt.Sprintf("OBJ.GROUP.{%d}", groupID))
	}
}

func (r *repoGroups) updateParticipantsCount(groupID int64) {
	count := int32(0)
	_ = r.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = r.getPrefix(groupID)
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		for it.Seek(r.getGroupParticipantKey(groupID, 0)); it.ValidForPrefix(opts.Prefix); it.Next() {
			count++
		}
		it.Close()
		return nil
	})

	group := r.Get(groupID)
	group.Participants = count
	r.Save(group)
}

func (r *repoGroups) Save(group *msg.Group) {
	if alreadySaved(fmt.Sprintf("G.%d", group.ID), group) {
		return
	}
	defer r.deleteFromCache(group.ID)

	if group == nil {
		return
	}

	groupKey := r.getGroupKey(group.ID)
	groupBytes, _ := group.Marshal()
	_ = r.badger.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(
			groupKey, groupBytes,
		))
	})

	_ = r.searchIndex.Index(ronak.ByteToStr(groupKey), GroupSearch{
		Type:   "group",
		Title:  group.Title,
		PeerID: group.ID,
	})
}

func (r *repoGroups) SaveMany(groups []*msg.Group) {

	groupIDs := domain.MInt64B{}
	for _, v := range groups {
		if alreadySaved(fmt.Sprintf("G.%d", v.ID), v) {
			continue
		}
		groupIDs[v.ID] = true
	}
	defer r.deleteFromCache(groupIDs.ToArray()...)

	for idx := range groups {
		r.Save(groups[idx])
	}

	return
}

func (r *repoGroups) SaveParticipant(groupID int64, participant *msg.GroupParticipant) {

	defer r.deleteFromCache(groupID)

	if participant == nil {
		return
	}

	groupParticipantKey := r.getGroupParticipantKey(groupID, participant.UserID)
	participantBytes, _ := participant.Marshal()
	_ = r.badger.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(
			groupParticipantKey, participantBytes,
		))
	})
}

func (r *repoGroups) GetMany(groupIDs []int64) []*msg.Group {
	return r.readManyFromCache(groupIDs)
}

func (r *repoGroups) Get(groupID int64) *msg.Group {
	return r.readFromCache(groupID)
}

func (r *repoGroups) GetParticipant(groupID int64, memberID int64) *msg.GroupParticipant {

	gp := new(msg.GroupParticipant)
	_ = r.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get(r.getGroupParticipantKey(groupID, memberID))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return gp.Unmarshal(val)
		})
	})
	return gp
}

func (r *repoGroups) GetParticipants(groupID int64) ([]*msg.GroupParticipant, error) {

	participants := make([]*msg.GroupParticipant, 0, 100)
	_ = r.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = r.getPrefix(groupID)
		it := txn.NewIterator(opts)
		for it.Seek(r.getGroupParticipantKey(groupID, 0)); it.ValidForPrefix(opts.Prefix); it.Next() {
			p := new(msg.GroupParticipant)
			_ = it.Item().Value(func(val []byte) error {
				return p.Unmarshal(val)
			})
			participants = append(participants, p)
		}
		it.Close()
		return nil
	})

	return participants, nil
}

func (r *repoGroups) DeleteMember(groupID, userID int64) {

	group := r.Get(groupID)
	if group == nil {
		return
	}

	_ = r.badger.Update(func(txn *badger.Txn) error {
		return txn.Delete(r.getGroupParticipantKey(groupID, userID))
	})

	r.updateParticipantsCount(groupID)
}

func (r *repoGroups) DeleteAllMembers(groupID int64) {

	_ = r.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = r.getPrefix(groupID)
		it := txn.NewIterator(opts)
		for it.Seek(r.getGroupParticipantKey(groupID, 0)); it.ValidForPrefix(opts.Prefix); it.Next() {
			_ = txn.Delete(it.Item().Key())
		}
		it.Close()
		return nil
	})

	group := r.Get(groupID)
	group.Participants = 0
	r.Save(group)
	return
}

func (r *repoGroups) UpdateTitle(groupID int64, title string) {

	defer r.deleteFromCache(groupID)

	group := r.Get(groupID)
	group.Title = title
	r.Save(group)
}

func (r *repoGroups) DeleteMemberMany(groupID int64, memberIDs []int64) {

	for _, memberID := range memberIDs {
		r.DeleteMember(groupID, memberID)
	}
}

func (r *repoGroups) Delete(groupID int64) {

	defer r.deleteFromCache(groupID)

	_ = r.badger.Update(func(txn *badger.Txn) error {
		err := txn.Delete(r.getGroupKey(groupID))
		if err != nil {
			return err
		}
		r.DeleteAllMembers(groupID)
		return nil
	})
}

func (r *repoGroups) UpdateMemberType(groupID, userID int64, isAdmin bool) {

	defer r.deleteFromCache(groupID)

	group := r.Get(groupID)
	if group == nil {
		return
	}
	flags := make([]msg.GroupFlags, 0, len(group.Flags))
	for _, f := range group.Flags {
		if f != msg.GroupFlagsAdmin {
			flags = append(flags, f)
		}
	}
	gp := r.GetParticipant(groupID, userID)
	if isAdmin {
		flags = append(flags, msg.GroupFlagsAdmin)
		gp.Type = msg.ParticipantTypeAdmin
	} else {
		gp.Type = msg.ParticipantTypeMember
	}
	group.Flags = flags
	r.SaveParticipant(groupID, gp)
	r.Save(group)
}

func (r *repoGroups) RemovePhoto(groupID int64) {

	defer r.deleteFromCache(groupID)

	group := r.Get(groupID)
	if group == nil {
		return
	}
	group.Photo = nil
	r.Save(group)
}

func (r *repoGroups) UpdatePhoto(groupPhoto *msg.UpdateGroupPhoto) {
	if alreadySaved(fmt.Sprintf("GPHOTO.%d", groupPhoto.GroupID), groupPhoto) {
		return
	}

	defer r.deleteFromCache(groupPhoto.GroupID)

	group := r.Get(groupPhoto.GroupID)
	group.Photo = groupPhoto.Photo
	r.Save(group)
}

func (r *repoGroups) Search(searchPhrase string) []*msg.Group {
	t1 := bleve.NewTermQuery("group")
	t1.SetField("type")
	terms := strings.Fields(searchPhrase)
	qs := make([]query.Query, 0)
	for _, term := range terms {
		qs = append(qs, bleve.NewPrefixQuery(term), bleve.NewMatchQuery(term), bleve.NewFuzzyQuery(term))
	}
	t2 := bleve.NewDisjunctionQuery(qs...)
	searchRequest := bleve.NewSearchRequest(bleve.NewConjunctionQuery(t1, t2))
	searchResult, _ := r.searchIndex.Search(searchRequest)
	groups := make([]*msg.Group, 0, 100)
	for _, hit := range searchResult.Hits {
		group := r.getGroupByKey(ronak.StrToByte(hit.ID))
		if group != nil {
			groups = append(groups, group)
		}
	}
	return groups
}
