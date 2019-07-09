package repo

import (
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"
	"go.uber.org/zap"
)

type repoGroups struct {
	*repository
}

func (r *repoGroups) readFromDb(groupID int64) *msg.Group {
	r.mx.Lock()
	defer r.mx.Unlock()

	pbGroup := new(msg.Group)
	groups := new(dto.Groups)

	err := r.db.Find(groups, groupID).Error
	if err != nil {
		logs.Error("Groups::GetGroup()-> fetch groups entity", zap.Error(err))
		return nil
	}

	groups.MapTo(pbGroup)

	return pbGroup
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

func (r *repoGroups) Save(g *msg.Group) (err error) {
	if alreadySaved(fmt.Sprintf("G.%d", g.ID), g) {
		return nil
	}
	defer r.deleteFromCache(g.ID)
	r.mx.Lock()
	defer r.mx.Unlock()

	ge := new(dto.Groups)
	r.db.Find(ge, g.ID)
	isNew := ge.ID == 0
	ge.MapFrom(g)
	if isNew {
		return r.db.Create(ge).Error
	}
	return r.db.Table(ge.TableName()).Where("ID=?", g.ID).Update(ge).Error
}

func (r *repoGroups) SaveMany(groups []*msg.Group) error {
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
	mapDTOGroups := make(map[int64]*dto.Groups)
	dtoGroups := make([]dto.Groups, 0)
	err := r.db.Where("ID in (?)", groupIDs.ToArray()).Find(&dtoGroups).Error
	if err != nil {
		logs.Error("Groups::SaveMany()-> fetch groups entity", zap.Error(err))
		return err
	}
	count := len(dtoGroups)
	for i := 0; i < count; i++ {
		mapDTOGroups[dtoGroups[i].ID] = &dtoGroups[i]
	}

	for _, v := range groups {
		if dtoEntity, ok := mapDTOGroups[v.ID]; ok {
			dtoEntity.MapFrom(v)
			err = r.db.Table(dtoEntity.TableName()).Where("ID=?", dtoEntity.ID).Update(dtoEntity).Error
		} else {
			dtoEntity := new(dto.Groups)
			dtoEntity.MapFrom(v)
			err = r.db.Create(dtoEntity).Error
		}
		if err != nil {
			logs.Error("Groups::SaveMany()-> save group entity",
				zap.Int64("ID", v.ID),
				zap.String("Title", v.Title),
				zap.Int64("CreatedOn", v.CreatedOn),
				zap.Error(err))
			break
		}
	}
	return err
}

func (r *repoGroups) GetMany(groupIDs []int64) []*msg.Group {
	return r.readManyFromCache(groupIDs)
}

func (r *repoGroups) Get(groupID int64) *msg.Group {
	return r.readFromCache(groupID)
}

func (r *repoGroups) DeleteMember(groupID, userID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	dtoGP := new(dto.GroupsParticipants)
	err := r.db.Where("GroupID = ? AND userID = ?", groupID, userID).First(dtoGP).Error
	if err == nil {
		err = r.db.Delete(dtoGP).Error
	}
	return err
}

func (r *repoGroups) DeleteAllMembers(groupID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	return r.db.Where("GroupID= ? ", groupID).Delete(dto.GroupsParticipants{}).Error
}

func (r *repoGroups) UpdateTitle(groupID int64, title string) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	defer r.deleteFromCache(groupID)
	dtoG := new(dto.Groups)
	return r.db.Table(dtoG.TableName()).Where("ID = ? ", groupID).Updates(map[string]interface{}{"Title": title}).Error
}

func (r *repoGroups) SaveParticipants(groupID int64, participant *msg.GroupParticipant) error {
	r.mx.Lock()
	defer r.mx.Unlock()
	defer r.deleteFromCache(groupID)
	dtoGP := new(dto.GroupsParticipants)
	err := r.db.Where("GroupID = ? AND userID = ?", groupID, participant.UserID).First(dtoGP).Error
	dtoGP.MapFrom(groupID, participant)
	// if record does not exist, not found error returns
	if err != nil {
		return r.db.Create(dtoGP).Error
	}
	return r.db.Table(dtoGP.TableName()).Where("GroupID = ? AND userID = ?", groupID, participant.UserID).Update(dtoGP).Error
}

func (r *repoGroups) GetParticipants(groupID int64) ([]*msg.GroupParticipant, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	dtoGPs := make([]dto.GroupsParticipants, 0)
	err := r.db.Where("GroupID = ?", groupID).Find(&dtoGPs).Error
	if err != nil {
		return nil, err
	}
	res := make([]*msg.GroupParticipant, 0)
	for _, v := range dtoGPs {
		tmp := new(msg.GroupParticipant)
		v.MapTo(tmp)
		res = append(res, tmp)
	}

	return res, nil
}

func (r *repoGroups) DeleteMemberMany(peerID int64, IDs []int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	return r.db.Where("GroupID= ? AND userID IN (?)", peerID, IDs).Delete(dto.GroupsParticipants{}).Error
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
		logs.Error("Groups::GetGroup()-> fetch groups entity", zap.Error(err))
		return nil, err // , err
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
