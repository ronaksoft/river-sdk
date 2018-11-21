package repo

import (
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo/dto"
	"go.uber.org/zap"
)

type RepoGroups interface {
	Save(g *msg.Group) (err error)
	GetManyGroups(groupIDs []int64) ([]*msg.Group, error)
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

// Get
func (r *repoGroups) GetManyGroups(groupIDs []int64) ([]*msg.Group, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	pbGroup := make([]*msg.Group, 0, len(groupIDs))
	groups := make([]dto.Groups, 0, len(groupIDs))

	err := r.db.Where("ID in (?)", groupIDs).Find(&groups).Error
	if err != nil {
		log.LOG_Debug("RepoGroups::GetManyGroups()-> fetch groups entity",
			zap.String("Error", err.Error()),
		)
		return nil, err
	}

	for idx, v := range groups {
		v.MapTo(pbGroup[idx])
	}

	return pbGroup, nil
}
