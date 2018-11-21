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
	SaveMany(groups []*msg.Group) error
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
	isNew := ge.ID <= 0
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

// Get
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
