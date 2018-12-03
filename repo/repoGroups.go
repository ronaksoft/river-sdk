package repo

import (
	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo/dto"
	"go.uber.org/zap"
)

type RepoGroups interface {
	Save(g *msg.Group) (err error)
	GetManyGroups(groupIDs []int64) []*msg.Group
	GetGroup(groupID int64) (*msg.Group, error)
	SaveMany(groups []*msg.Group) error
	// AddGroupMember(m *msg.UpdateGroupMemberAdded) error
	DeleteGroupMember(groupID, userID int64) error
	DeleteAllGroupMember(groupID int64) error
	UpdateGroupTitle(groupID int64, title string) error
	SaveParticipants(groupID int64, participant *msg.GroupParticipant) error
	GetParticipants(groupID int64) ([]*msg.GroupParticipant, error)
	DeleteGroupMemberMany(peerID int64, IDs []int64) error
	SaveParticipantsByID(groupID, createdOn int64, userIDs []int64)
	Delete(groupID int64) error
}

type repoGroups struct {
	*repository
}

// Get
func (r *repoGroups) Save(g *msg.Group) (err error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	ge := new(dto.Groups)
	r.db.Find(ge, g.ID)
	isNew := ge.ID == 0
	ge.MapFrom(g)
	if isNew {
		return r.db.Create(ge).Error
	} else {
		return r.db.Table(ge.TableName()).Where("ID=?", g.ID).Update(ge).Error
	}
}

func (r *repoGroups) SaveMany(groups []*msg.Group) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	groupIDs := domain.MInt64B{}
	for _, v := range groups {
		groupIDs[v.ID] = true
	}
	mapDTOGroups := make(map[int64]*dto.Groups)
	dtoGroups := make([]dto.Groups, 0)
	err := r.db.Where("ID in (?)", groupIDs.ToArray()).Find(&dtoGroups).Error
	if err != nil {
		log.LOG_Debug("RepoGroups::SaveMany()-> fetch groups entity",
			zap.String("Error", err.Error()),
		)
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
			log.LOG_Debug("RepoGroups::SaveMany()-> save group entity",
				zap.Int64("ID", v.ID),
				zap.String("Title", v.Title),
				zap.Int64("CreatedOn", v.CreatedOn),
				zap.String("Error", err.Error()),
			)
			break
		}
	}
	return err
}

func (r *repoGroups) GetManyGroups(groupIDs []int64) []*msg.Group {
	r.mx.Lock()
	defer r.mx.Unlock()

	pbGroup := make([]*msg.Group, 0)
	groups := make([]dto.Groups, 0, len(groupIDs))

	err := r.db.Where("ID in (?)", groupIDs).Find(&groups).Error
	if err != nil {
		log.LOG_Debug("RepoGroups::GetManyGroups()-> fetch groups entity",
			zap.String("Error", err.Error()),
		)
		return nil //, err
	}

	for _, v := range groups {
		tmp := new(msg.Group)
		v.MapTo(tmp)
		pbGroup = append(pbGroup, tmp)
	}

	return pbGroup
}

func (r *repoGroups) GetGroup(groupID int64) (*msg.Group, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	pbGroup := new(msg.Group)
	groups := new(dto.Groups)

	err := r.db.Find(groups, groupID).Error
	if err != nil {
		log.LOG_Debug("RepoGroups::GetGroup()-> fetch groups entity",
			zap.String("Error", err.Error()),
		)
		return nil, err //, err
	}

	groups.MapTo(pbGroup)

	return pbGroup, nil
}

// func (r *repoGroups) AddGroupMember(m *msg.UpdateGroupMemberAdded) error {
// 	r.mx.Lock()
// 	defer r.mx.Unlock()

// 	dtoGP := new(dto.GroupParticipants)
// 	err := r.db.Where("GroupID = ? AND UserID = ?", m.GroupID, m.UserID).First(dtoGP).Error
// 	// if record does not exist, not found error returns
// 	if err != nil {
// 		dtoGP.MapFromUpdateGroupMemberAdded(m)
// 		err = r.db.Create(dtoGP).Error
// 	}
// 	return err
// }

func (r *repoGroups) DeleteGroupMember(groupID, userID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	dtoGP := new(dto.GroupParticipants)
	err := r.db.Where("GroupID = ? AND UserID = ?", groupID, userID).First(dtoGP).Error
	if err == nil {
		err = r.db.Delete(dtoGP).Error
	}
	return err
}

func (r *repoGroups) DeleteAllGroupMember(groupID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	return r.db.Where("GroupID= ? ", groupID).Delete(dto.GroupParticipants{}).Error
}

func (r *repoGroups) UpdateGroupTitle(groupID int64, title string) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	dtoG := new(dto.Groups)
	return r.db.Table(dtoG.TableName()).Where("ID = ? ", groupID).Updates(map[string]interface{}{"Title": title}).Error
}

func (r *repoGroups) SaveParticipants(groupID int64, participant *msg.GroupParticipant) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	dtoGP := new(dto.GroupParticipants)
	err := r.db.Where("GroupID = ? AND UserID = ?", groupID, participant.UserID).First(dtoGP).Error
	dtoGP.MapFrom(groupID, participant)
	// if record does not exist, not found error returns
	if err != nil {
		return r.db.Create(dtoGP).Error
	}
	return r.db.Table(dtoGP.TableName()).Where("GroupID = ? AND UserID = ?", groupID, participant.UserID).Update(dtoGP).Error
}

func (r *repoGroups) GetParticipants(groupID int64) ([]*msg.GroupParticipant, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	dtoGPs := make([]dto.GroupParticipants, 0)
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

func (r *repoGroups) DeleteGroupMemberMany(peerID int64, IDs []int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	return r.db.Where("GroupID= ? AND UserID IN (?)", peerID, IDs).Delete(dto.GroupParticipants{}).Error
}

func (r *repoGroups) SaveParticipantsByID(groupID, createdOn int64, userIDs []int64) {
	r.mx.Lock()
	defer r.mx.Unlock()

	for _, id := range userIDs {
		dtoGP := new(dto.GroupParticipants)
		r.db.Where("GroupID = ? AND UserID = ?", groupID, id).First(dtoGP)
		dtoGP.Date = createdOn
		dtoGP.GroupID = groupID
		dtoGP.Type = int32(msg.ParticipantType_Member)
		dtoGP.UserID = id
		r.db.Save(dtoGP)
	}
}

func (r *repoGroups) Delete(groupID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	return r.db.Where("ID= ? ", groupID).Delete(dto.Groups{}).Error
}
