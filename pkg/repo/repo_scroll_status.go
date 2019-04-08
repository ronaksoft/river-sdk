package repo

import "git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"

// ScrollStatus repoScrollStatus interface
type ScrollStatus interface {
	Save(peerID, msgID int64) error
	Get(peerID int64) (int64, error)
}

type repoScrollStatus struct {
	*repository
}

// Save
func (r *repoScrollStatus) Save(peerID, msgID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	scrollStatus := dto.ScrollStatus{
		PeerID: peerID,
		MessageID: msgID,
	}
	s := new(dto.ScrollStatus)

	// check if scroll status for this peerID has been created before
	r.db.Table(scrollStatus.TableName()).Where("PeerID=?", peerID).Find(s)

	// create new entry in db
	if s.PeerID == 0 {
		return r.db.Create(scrollStatus).Error
	}

	// update existing
	return r.db.Table(scrollStatus.TableName()).Where("PeerID=?", peerID).Update(scrollStatus).Error
}

// Get
func (r *repoScrollStatus) Get(peerID int64) (int64, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	s := dto.ScrollStatus{}
	err := r.db.Table(s.TableName()).Where("PeerID=?", peerID).Find(&s).Error
	if err != nil {
		return 0, err
	} else {
		return s.MessageID, nil
	}
}