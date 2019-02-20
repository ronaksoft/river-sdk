package repo

import (
	"database/sql"
	"errors"

	"git.ronaksoftware.com/ronak/riversdk/repo/dto"
)

// MessageHoles repoMessageHoles interface
type MessageHoles interface {
	Save(peerID, minID, maxID int64) error
	Delete(peerID, minID int64) error
	DeleteAll(peerID int64) error
	GetHoles(peerID, msgMinID, msgMaxID int64) ([]dto.MessageHoles, error)
}

type repoMessageHoles struct {
	*repository
}

// Save
func (r *repoMessageHoles) Save(peerID, minID, maxID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if minID >= maxID {
		return errors.New("invalid hole params, minID is less or equal to maxID")
	}

	nilInt64 := sql.NullInt64{
		Int64: minID,
		Valid: true,
	}

	m := dto.MessageHoles{
		PeerID: peerID,
		MinID:  nilInt64,
		MaxID:  maxID,
	}

	entity := new(dto.MessageHoles)
	r.db.Table(m.TableName()).Where("PeerID = ? AND MinID=?", peerID, minID).Find(entity)
	if entity.PeerID == 0 {
		return r.db.Create(m).Error
	}
	return r.db.Table(m.TableName()).Where("PeerID = ? AND MinID=?", peerID, minID).Update(m).Error
}

// Delete
func (r *repoMessageHoles) Delete(peerID, minID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	return r.db.Where("PeerID = ? AND MinID=?", peerID, minID).Delete(dto.MessageHoles{}).Error
}

// DeleteAll
func (r *repoMessageHoles) DeleteAll(peerID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	return r.db.Where("PeerID = ?", peerID).Delete(dto.MessageHoles{}).Error
}

// GetHoles
func (r repoMessageHoles) GetHoles(peerID, msgMinID, msgMaxID int64) ([]dto.MessageHoles, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	/*
		A : MinID <= msgMinID && MinID < msgMaxID && MaxID > msgMinID && MaxID >= msgMaxID ||
		B : MinID > msgMinID && MinID < msgMaxID && MaxID > msgMinID && MaxID > msgMaxID ||
		C : MinID < msgMinID && MinID < msgMaxID && MaxID > msgMinID && MaxID < msgMaxID ||
		D : MinID >= msgMinID && MinID < msgMaxID && MaxID > msgMinID && MaxID <= msgMaxID ||
	*/
	dtoHoles := make([]dto.MessageHoles, 0)
	err := r.db.Where("PeerID = ? AND ("+
		"(MinID <= ? AND MinID < ? AND MaxID > ? AND MaxID >= ?)"+" OR "+ // A : inside or exact size of hole
		"(MinID > ? AND MinID < ? AND MaxID > ? AND MaxID > ?)"+" OR "+ // B : minside overlap
		"(MinID < ? AND MinID < ? AND MaxID > ? AND MaxID < ?)"+" OR "+ // C : maxside overlap
		"(MinID >= ? AND MinID < ? AND MaxID > ? AND MaxID <= ?)"+" ) ", // D : surrendered over hole
		peerID,
		msgMinID, msgMaxID, msgMinID, msgMaxID,
		msgMinID, msgMaxID, msgMinID, msgMaxID,
		msgMinID, msgMaxID, msgMinID, msgMaxID,
		msgMinID, msgMaxID, msgMinID, msgMaxID,
	).Order("PeerID, MinID, MaxID").Find(&dtoHoles).Error
	return dtoHoles, err
}
