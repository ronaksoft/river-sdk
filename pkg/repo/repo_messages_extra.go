package repo

import "git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"

// MessagesExtra repoScrollStatus interface
type MessagesExtra interface {
	SaveScrollID(peerID, msgID int64, peerType int32) error
	GetScrollID(peerID int64, peerType int32) (int64, error)
	DeleteScrollID(peerID int64, peerType int32) error
}

type repoMessagesExtra struct {
	*repository
}

// Save
func (r *repoMessagesExtra) SaveScrollID(peerID, msgID int64, peerType int32) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	scrollStatus := dto.MessagesExtra{
		PeerID:   peerID,
		PeerType: peerType,
		ScrollID: msgID,
	}
	s := new(dto.MessagesExtra)

	// check if scroll status for this peerID has been created before
	r.db.Table(scrollStatus.TableName()).Where("PeerID=? AND PeerType=?", peerID, peerType).Find(s)

	// create new entry in db
	if s.PeerID == 0 {
		return r.db.Create(scrollStatus).Error
	}

	// update existing
	return r.db.Table(scrollStatus.TableName()).Where("PeerID=? AND PeerType=?", peerID, peerType).Update(scrollStatus).Error
}

// Get
func (r *repoMessagesExtra) GetScrollID(peerID int64, peerType int32) (int64, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	s := dto.MessagesExtra{}
	err := r.db.Table(s.TableName()).Where("PeerID=? AND PeerType=?", peerID, peerType).Find(&s).Error
	if err != nil {
		return 0, err
	} else {
		return s.ScrollID, nil
	}
}

// Delete
func (r *repoMessagesExtra) DeleteScrollID(peerID int64, peerType int32) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	// Delete Row
	return r.db.Where("PeerID=? AND  PeerType=?", peerID, peerType).Delete(dto.MessagesExtra{}).Error
}
