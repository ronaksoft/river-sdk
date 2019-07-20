package repo

import (
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/dgraph-io/badger"
	"go.uber.org/zap"
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
func (r *repoUsers) getGroupByKey(groupKey []byte) *msg.Group {
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
		it := txn.NewIterator(opts)
		for it.Seek(r.getGroupParticipantKey(groupID, 0)); it.Valid(); it.Next() {
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
	r.mx.Lock()
	defer r.mx.Unlock()

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

	_ = r.searchIndex.Index(ronak.ByteToStr(groupKey), Group{
		Type:   "user",
		Title:  group.Title,
		PeerID: group.ID,
	})
}

func (r *repoGroups) SaveMany(groups []*msg.Group) {
	r.mx.Lock()
	defer r.mx.Unlock()

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

func (r *repoGroups) SaveParticipants(groupID int64, participant *msg.GroupParticipant) {
	r.mx.Lock()
	defer r.mx.Unlock()
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

func (r *repoGroups) DeleteMember(groupID, userID int64) {
	r.mx.Lock()
	defer r.mx.Unlock()

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
	r.mx.Lock()
	defer r.mx.Unlock()

	_ = r.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = r.getPrefix(groupID)
		it := txn.NewIterator(opts)
		for it.Seek(r.getGroupParticipantKey(groupID, 0)); it.Valid(); it.Next() {
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

func (r *repoGroups) UpdateTitle(groupID int64, title string)  {
	r.mx.Lock()
	defer r.mx.Unlock()
	defer r.deleteFromCache(groupID)

	group := r.Get(groupID)
	group.Title = title
	r.Save(group)
}

func (r *repoGroups) GetParticipants(groupID int64) ([]*msg.GroupParticipant, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	participants := make([]*msg.GroupParticipant, 0, 100)
	_ = r.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = r.getPrefix(groupID)
		it := txn.NewIterator(opts)
		for it.Seek(r.getGroupParticipantKey(groupID, 0)); it.Valid(); it.Next() {
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

func (r *repoGroups) DeleteMemberMany(groupID int64, memberIDs []int64) {
	r.mx.Lock()
	defer r.mx.Unlock()

	for _, memberID := range memberIDs {
		r.DeleteMember(groupID, memberID)
	}
}

func (r *repoGroups) Delete(groupID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	return r.db.Where("ID= ? ", groupID).Delete(dto.Groups{}).Error
}

func (r *repoGroups) UpdateMemberType(groupID, userID int64, isAdmin bool) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	defer r.deleteFromCache(groupID)
	dtoGP := new(dto.GroupsParticipants)

	userType := int32(msg.ParticipantTypeMember)
	if isAdmin {
		userType = int32(msg.ParticipantTypeAdmin)
	}

	return r.db.Table(dtoGP.TableName()).Where("GroupID = ? AND userID = ?", groupID, userID).Updates(map[string]interface{}{
		"Type": userType,
	}).Error
}

func (r *repoGroups) Search(searchPhrase string) []*msg.Group {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("Groups::SearchGroups()")

	p := "%" + searchPhrase + "%"
	users := make([]dto.Groups, 0)
	err := r.db.Where("Title LIKE ? ", p).Find(&users).Error
	if err != nil {
		logs.Error("Groups::SearchGroups()-> fetch group entities", zap.Error(err))
		return nil
	}
	pbGroup := make([]*msg.Group, 0)
	for _, v := range users {
		tmpG := new(msg.Group)
		v.MapTo(tmpG)
		pbGroup = append(pbGroup, tmpG)
	}

	return pbGroup
}

func (r *repoGroups) GetGroupDTO(groupID int64) (*dto.Groups, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	group := new(dto.Groups)

	err := r.db.Find(group, groupID).Error
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (r *repoGroups) UpdatePhotoPath(groupID int64, isBig bool, filePath string) error {
	r.mx.Lock()
	defer r.mx.Unlock()
	defer r.deleteFromCache(groupID)
	e := new(dto.Groups)

	if isBig {
		return r.db.Table(e.TableName()).Where("ID = ? ", groupID).Updates(map[string]interface{}{
			"BigFilePath": filePath,
		}).Error
	}

	return r.db.Table(e.TableName()).Where("ID = ? ", groupID).Updates(map[string]interface{}{
		"SmallFilePath": filePath,
	}).Error

}

func (r *repoGroups) UpdatePhoto(groupPhoto *msg.UpdateGroupPhoto) error {
	if alreadySaved(fmt.Sprintf("GPHOTO.%d", groupPhoto.GroupID), groupPhoto) {
		return nil
	}
	r.mx.Lock()
	defer r.mx.Unlock()
	defer r.deleteFromCache(groupPhoto.GroupID)
	grp := new(dto.Groups)
	err := r.db.Find(grp, groupPhoto.GroupID).Error
	if err == nil {
		grp.MapFromUpdateGroupPhoto(groupPhoto)
		return r.db.Table(grp.TableName()).Where("ID=?", grp.ID).Updates(grp).Error
	}
	return err
}

func (r *repoGroups) RemovePhoto(groupID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()
	defer r.deleteFromCache(groupID)
	grp := new(dto.Groups)
	return r.db.Table(grp.TableName()).Where("ID=?", groupID).Updates(map[string]interface{}{
		"Photo":           []byte("[]"),
		"BigFileID":       0,
		"BigAccessHash":   0,
		"BigClusterID":    0,
		"BigVersion":      0,
		"BigFilePath":     "",
		"SmallFileID":     0,
		"SmallAccessHash": 0,
		"SmallClusterID":  0,
		"SmallVersion":    0,
		"SmallFilePath":   "",
	}).Error
}

func (r *repoGroups) SearchByTitle(title string) []*msg.Group {
	r.mx.Lock()
	defer r.mx.Unlock()

	pbGroup := make([]*msg.Group, 0)
	groups := make([]dto.Groups, 0)

	err := r.db.Where("Title LIKE ?", "%"+fmt.Sprintf("%s", title)+"%").Find(&groups).Error
	if err != nil {
		logs.Error("Groups::SearchGroupsByTitle()-> fetch groups entity", zap.Error(err))
		return nil
	}

	for _, v := range groups {
		tmp := new(msg.Group)
		v.MapTo(tmp)
		pbGroup = append(pbGroup, tmp)
	}

	return pbGroup
}
