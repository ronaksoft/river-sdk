package repo

import (
	"fmt"
	msg "git.ronaksoftware.com/river/msg/chat"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"github.com/dgraph-io/badger"
	"go.uber.org/zap"
	"strings"
	"time"
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

func getGroupParticipantPrefix(groupID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%021d.", prefixGroupsParticipants, groupID))
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

	indexPeer(
		ronak.ByteToStr(groupKey),
		GroupSearch{
			Type:   "group",
			Title:  group.Title,
			PeerID: group.ID,
		},
	)

	err = saveGroupPhotos(txn, group.ID, group.Photo)
	if err != nil {
		return err
	}
	return nil
}

func removeGroupPhotoGallery(txn *badger.Txn, groupID int64, photoIDs ...int64) error {
	for _, photoID := range photoIDs {
		err := txn.Delete(getGroupPhotoGalleryKey(groupID, photoID))
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
	}
	return nil
}

func updateGroupParticipantsCount(txn *badger.Txn, group *msg.Group) error {
	count := int32(0)
	opts := badger.DefaultIteratorOptions
	opts.Prefix = getGroupPrefix(group.ID)
	opts.PrefetchValues = false
	it := txn.NewIterator(opts)
	for it.Seek(getGroupParticipantKey(group.ID, 0)); it.ValidForPrefix(opts.Prefix); it.Next() {
		count++
	}
	it.Close()
	group.Participants = count
	return saveGroup(txn, group)
}

func (r *repoGroups) Save(groups ...*msg.Group) {
	groupIDs := domain.MInt64B{}
	for _, v := range groups {
		groupIDs[v.ID] = true
	}

	_ = badgerUpdate(func(txn *badger.Txn) error {
		for _, group := range groups {
			err := saveGroup(txn, group)
			logs.WarnOnErr("RepoGroups got error on save", err, zap.Int64("GroupID", group.ID))
		}
		return nil
	})
	return
}

func (r *repoGroups) GetMany(groupIDs []int64) []*msg.Group {
	groups := make([]*msg.Group, 0, len(groupIDs))
	_ = badgerView(func(txn *badger.Txn) error {
		for _, groupID := range groupIDs {
			if groupID == 0 {
				continue
			}
			group, err := getGroupByKey(txn, getGroupKey(groupID))
			logs.WarnOnErr("RepoGroups got error on get many", err, zap.Int64("GroupID", groupID))
			if group != nil {
				groups = append(groups, group)
			}
		}
		return nil
	})
	return groups
}

func (r *repoGroups) Get(groupID int64) (group *msg.Group, err error) {
	err = badgerView(func(txn *badger.Txn) error {
		group, err = getGroupByKey(txn, getGroupKey(groupID))
		if err != nil {
			return err
		}
		return nil
	})
	logs.WarnOnErr("RepoGroups got error on get", err)
	return
}

func (r *repoGroups) Delete(groupID int64) {
	_ = badgerUpdate(func(txn *badger.Txn) error {
		err := txn.Delete(getGroupKey(groupID))
		if err != nil {
			return err
		}
		r.DeleteAllMembers(groupID)
		return nil
	})
}

func (r *repoGroups) SaveParticipant(groupID int64, participant *msg.GroupParticipant) {
	if participant == nil {
		return
	}

	groupParticipantKey := getGroupParticipantKey(groupID, participant.UserID)
	participantBytes, _ := participant.Marshal()
	_ = badgerUpdate(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(
			groupParticipantKey, participantBytes,
		))
	})
}

func (r *repoGroups) GetParticipant(groupID int64, memberID int64) *msg.GroupParticipant {
	gp := new(msg.GroupParticipant)
	_ = badgerView(func(txn *badger.Txn) error {
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
	err := badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = getGroupParticipantPrefix(groupID)
		it := txn.NewIterator(opts)
		for it.Rewind(); it.Valid(); it.Next() {
			p := new(msg.GroupParticipant)
			_ = it.Item().Value(func(val []byte) error {
				return p.Unmarshal(val)
			})
			participants = append(participants, p)
		}
		it.Close()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return participants, nil
}

func (r *repoGroups) UpdatePhoto(groupID int64, groupPhoto *msg.GroupPhoto) {
	err := badgerUpdate(func(txn *badger.Txn) error {
		group, err := getGroupByKey(txn, getGroupKey(groupID))
		if err != nil {
			return err
		}
		group.Photo = groupPhoto
		return saveGroup(txn, group)
	})
	logs.WarnOnErr("RepoGroups got error on update photo", err)
}

func (r *repoGroups) RemovePhoto(groupID int64) {
	err := badgerUpdate(func(txn *badger.Txn) error {
		group, err := getGroupByKey(txn, getGroupKey(groupID))
		if err != nil {
			return err
		}
		group.Photo = nil

		err = removeGroupPhotoGallery(txn, groupID, group.Photo.PhotoID)
		if err != nil {
			return err
		}
		return saveGroup(txn, group)
	})
	logs.WarnOnErr("RepoGroups got error on update photo", err)
}

func (r *repoGroups) SavePhotoGallery(groupID int64, photos ...*msg.GroupPhoto) {
	err := badgerUpdate(func(txn *badger.Txn) error {
		return saveGroupPhotos(txn, groupID, photos...)
	})
	logs.ErrorOnErr("RepoGroups got error on save photo gallery", err)
}

func (r *repoGroups) RemovePhotoGallery(groupID int64, photoIDs ...int64) {
	err := badgerUpdate(func(txn *badger.Txn) error {
		return removeGroupPhotoGallery(txn, groupID, photoIDs...)
	})
	logs.WarnOnErr("RepoGroups got error on remove photo gallery", err)
}

func (r *repoGroups) GetPhotoGallery(groupID int64) []*msg.GroupPhoto {
	photos := make([]*msg.GroupPhoto, 0, 5)
	_ = badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = getGroupPhotoGalleryPrefix(groupID)
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
	err := badgerUpdate(func(txn *badger.Txn) error {
		group, err := getGroupByKey(txn, getGroupKey(groupID))
		if err != nil {
			return err
		}
		return updateGroupParticipantsCount(txn, group)

	})
	logs.ErrorOnErr("RepoGroups got error on delete member", err)
}

func (r *repoGroups) DeleteAllMembers(groupID int64) {
	err := badgerUpdate(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = getGroupParticipantPrefix(groupID)
		it := txn.NewIterator(opts)
		for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
			_ = txn.Delete(it.Item().KeyCopy(nil))
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
	err := badgerUpdate(func(txn *badger.Txn) error {
		group, err := getGroupByKey(txn, getGroupKey(groupID))
		if err != nil {
			return err
		}
		group.Title = title
		return saveGroup(txn, group)
	})
	logs.ErrorOnErr("RepoGroups got error on update photo", err)
}

func (r *repoGroups) DeleteMemberMany(groupID int64, memberIDs []int64) {
	for _, memberID := range memberIDs {
		r.DeleteMember(groupID, memberID)
	}
}

func (r *repoGroups) UpdateMemberType(groupID, userID int64, isAdmin bool) {
	err := badgerUpdate(func(txn *badger.Txn) error {
		group, err := getGroupByKey(txn, getGroupKey(groupID))
		if err != nil {
			return err
		}
		flags := make([]msg.GroupFlags, 0, len(group.Flags))
		for _, f := range group.Flags {
			if f != msg.GroupFlagsAdmin {
				flags = append(flags, f)
			}
		}
		gp := new(msg.GroupParticipant)
		item, err := txn.Get(getGroupParticipantKey(groupID, userID))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			return gp.Unmarshal(val)
		})
		if err != nil {
			return err
		}
		if isAdmin {
			flags = append(flags, msg.GroupFlagsAdmin)
			gp.Type = msg.ParticipantTypeAdmin
		} else {
			gp.Type = msg.ParticipantTypeMember
		}
		group.Flags = flags
		groupParticipantKey := getGroupParticipantKey(groupID, gp.UserID)
		participantBytes, _ := gp.Marshal()
		err = txn.SetEntry(badger.NewEntry(
			groupParticipantKey, participantBytes,
		))
		if err != nil {
			return err
		}
		return saveGroup(txn, group)
	})
	logs.WarnOnErr("RepoGroups got error on update member type", err)
}

func (r *repoGroups) Search(searchPhrase string) []*msg.Group {
	groups := make([]*msg.Group, 0, 100)
	if r.peerSearch == nil {
		return groups
	}
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
	_ = badgerView(func(txn *badger.Txn) error {
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
	err := ronak.Try(10, time.Second, func() error {
		if r.peerSearch == nil {
			return domain.ErrDoesNotExists
		}
		return nil
	})
	if err != nil {
		return
	}
	err = badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = ronak.StrToByte(prefixGroups)
		it := txn.NewIterator(opts)
		for it.Rewind(); it.Valid(); it.Next() {
			_ = it.Item().Value(func(val []byte) error {
				group := new(msg.Group)
				_ = group.Unmarshal(val)
				groupKey := ronak.ByteToStr(getGroupKey(group.ID))
				if d, _ := r.peerSearch.Document(groupKey); d == nil {
					indexPeer(
						groupKey,
						GroupSearch{
							Type:   "group",
							Title:  group.Title,
							PeerID: group.ID,
						},
					)
				}
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
