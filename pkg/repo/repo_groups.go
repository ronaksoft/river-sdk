package repo

import (
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"github.com/dgraph-io/badger"
	"go.uber.org/zap"
	"strings"
)

const (
	prefixGroups             = "GRP"
	prefixGroupsParticipants = "GRP_P"
	prefixGroupsPhotoGallery = "GRP_PHG"
)

type repoGroups struct {
	*repository
}

func getGroupKey(groupID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%021d", prefixGroups, groupID))
}

func getGroupByKey(txn *badger.Txn, groupKey []byte) (*msg.Group, error) {
	group := &msg.Group{}
	item, err := txn.Get(groupKey)
	if err != nil {
		return nil, err
	}
	err = item.Value(func(val []byte) error {
		return group.Unmarshal(val)
	})
	if err != nil {
		return nil, err
	}
	return group, nil
}

func getGroupParticipantKey(groupID, memberID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%021d.%021d", prefixGroupsParticipants, groupID, memberID))
}

func getGroupPhotoGalleryKey(groupID, photoID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%021d.%021d", prefixGroupsPhotoGallery, groupID, photoID))
}

func getGroupPhotoGalleryPrefix(groupID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%021d.", prefixGroupsPhotoGallery, groupID))
}

func getGroupPrefix(groupID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%021d.", prefixGroupsParticipants, groupID))
}

func saveGroup(txn *badger.Txn, group *msg.Group) error {
	groupKey := getGroupKey(group.ID)
	groupBytes, _ := group.Marshal()
	err := txn.SetEntry(badger.NewEntry(
		groupKey, groupBytes,
	))
	if err != nil {
		return err
	}

	_ = r.peerSearch.Index(ronak.ByteToStr(groupKey), GroupSearch{
		Type:   "group",
		Title:  group.Title,
		PeerID: group.ID,
	})

	err = saveGroupPhotos(txn, group.ID, group.Photo)
	if err != nil {
		return err
	}
	return nil
}

func (r *repoGroups) readFromDb(groupID int64) *msg.Group {
	group := new(msg.Group)
	err := r.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get(getGroupKey(groupID))
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
		opts.Prefix = getGroupPrefix(groupID)
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		for it.Seek(getGroupParticipantKey(groupID, 0)); it.ValidForPrefix(opts.Prefix); it.Next() {
			count++
		}
		it.Close()
		return nil
	})

	group := r.Get(groupID)
	group.Participants = count
	r.Save(group)
}

func (r *repoGroups) Save(groups ...*msg.Group) {
	groupIDs := domain.MInt64B{}
	for _, v := range groups {
		if alreadySaved(fmt.Sprintf("G.%d", v.ID), v) {
			continue
		}
		groupIDs[v.ID] = true
	}
	defer r.deleteFromCache(groupIDs.ToArray()...)

	for idx := range groups {
		r.save(groups[idx])
	}

	return
}
func (r *repoGroups) save(group *msg.Group) {

}

func (r *repoGroups) GetMany(groupIDs []int64) []*msg.Group {
	return r.readManyFromCache(groupIDs)
}

func (r *repoGroups) Get(groupID int64) *msg.Group {
	return r.readFromCache(groupID)
}

func (r *repoGroups) Delete(groupID int64) {
	defer r.deleteFromCache(groupID)

	_ = r.badger.Update(func(txn *badger.Txn) error {
		err := txn.Delete(getGroupKey(groupID))
		if err != nil {
			return err
		}
		r.DeleteAllMembers(groupID)
		return nil
	})
}

func (r *repoGroups) SaveParticipant(groupID int64, participant *msg.GroupParticipant) {
	defer r.deleteFromCache(groupID)
	if participant == nil {
		return
	}

	groupParticipantKey := getGroupParticipantKey(groupID, participant.UserID)
	participantBytes, _ := participant.Marshal()
	_ = r.badger.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(
			groupParticipantKey, participantBytes,
		))
	})
}

func (r *repoGroups) GetParticipant(groupID int64, memberID int64) *msg.GroupParticipant {
	gp := new(msg.GroupParticipant)
	_ = r.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get(getGroupParticipantKey(groupID, memberID))
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
		opts.Prefix = getGroupPrefix(groupID)
		it := txn.NewIterator(opts)
		for it.Seek(getGroupParticipantKey(groupID, 0)); it.ValidForPrefix(opts.Prefix); it.Next() {
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

func (r *repoGroups) UpdatePhoto(groupID int64, groupPhoto *msg.GroupPhoto) {
	if alreadySaved(fmt.Sprintf("GPHOTO.%d", groupID), groupPhoto) {
		return
	}

	defer r.deleteFromCache(groupID)

	group := r.Get(groupID)
	group.Photo = groupPhoto
	r.Save(group)
}

func (r *repoGroups) RemovePhoto(groupID int64) {
	defer r.deleteFromCache(groupID)

	group := r.Get(groupID)
	if group == nil {
		return
	}

	// 1. Remove the photo from the photo gallery of the user
	r.RemovePhotoGallery(groupID, group.Photo.PhotoID)

	// 2. Save Group object into the db again
	group.Photo = nil
	r.Save(group)
}

func (r *repoGroups) SavePhotoGallery(groupID int64, photos ...*msg.GroupPhoto) {
	err := r.badger.Update(func(txn *badger.Txn) error {
		return saveGroupPhotos(txn, groupID, photos...)
	})
	logs.ErrorOnErr("RepoGroups got error on save photo gallery", err)
}

func (r *repoGroups) RemovePhotoGallery(groupID int64, photoIDs ...int64) {
	_ = r.badger.Update(func(txn *badger.Txn) error {
		for _, photoID := range photoIDs {
			_  =txn.Delete(getGroupPhotoGalleryKey(groupID, photoID))
		}
		return nil
	})
}

func (r *repoGroups) GetPhotoGallery(groupID int64) []*msg.GroupPhoto {
	photos := make([]*msg.GroupPhoto, 0, 5)
	_ = r.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix =getGroupPhotoGalleryPrefix(groupID)
		it := txn.NewIterator(opts)
		for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
			_ = it.Item().Value(func(val []byte) error {
				groupPhoto := new(msg.GroupPhoto)
				err := groupPhoto.Unmarshal(val)
				if err != nil {
					return err
				}
				photos = append(photos, groupPhoto)
				return nil
			})
		}
		it.Close()
		return nil
	})
	return photos
}

func (r *repoGroups) DeleteMember(groupID, userID int64) {
	group := r.Get(groupID)
	if group == nil {
		return
	}

	_ = r.badger.Update(func(txn *badger.Txn) error {
		return txn.Delete(getGroupParticipantKey(groupID, userID))
	})

	r.updateParticipantsCount(groupID)
}

func (r *repoGroups) DeleteAllMembers(groupID int64) {
	err := r.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = getGroupPrefix(groupID)
		it := txn.NewIterator(opts)
		for it.Seek(getGroupParticipantKey(groupID, 0)); it.ValidForPrefix(opts.Prefix); it.Next() {
			_ = txn.Delete(it.Item().Key())
		}
		it.Close()

		group, err := getGroupByKey(txn, getGroupKey(groupID))
		if err != nil {
			return err
		}
		group.Participants = 0
		return saveGroup(txn, group)
	})
	logs.ErrorOnErr("RepoGroup got error on delete all members", err)
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
	searchResult, _ := r.peerSearch.Search(searchRequest)
	groups := make([]*msg.Group, 0, 100)
	_ = r.badger.View(func(txn *badger.Txn) error {
		for _, hit := range searchResult.Hits {
			group, _ := getGroupByKey(txn, ronak.StrToByte(hit.ID))
			if group != nil {
				groups = append(groups, group)
			}
		}
		return nil
	})

	return groups
}

func (r *repoGroups) ReIndex() {
	err := r.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = ronak.StrToByte(prefixGroups)
		it := txn.NewIterator(opts)
		for it.Rewind(); it.Valid(); it.Next() {
			_ = it.Item().Value(func(val []byte) error {
				group := new(msg.Group)
				_ = group.Unmarshal(val)
				groupKey := getGroupKey(group.ID)
				_ = r.peerSearch.Index(ronak.ByteToStr(groupKey), GroupSearch{
					Type:   "group",
					Title:  group.Title,
					PeerID: group.ID,
				})
				return nil
			})
		}
		it.Close()
		return nil
	})
	if err != nil {
		logs.Warn("Error On ReIndex Users", zap.Error(err))
	}
}
