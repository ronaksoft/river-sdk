package repo

import (
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo/dto"
)

type RepoFiles interface {
	SaveFileStatus(fileID int64, filePath string, position, totalSize int64, partNo, totalParts int32) (err error)
	SaveFileMessageMedia(fileID int64, req *msg.ClientSendMessageMedia) error
	GetDialogFileMessageMedia(peerID int64, peerType int32) ([]*msg.ClientSendMessageMedia, error)
}

type repoFiles struct {
	*repository
}

// SaveFileStatus
func (r *repoFiles) SaveFileStatus(fileID int64, filePath string, position, totalSize int64, partNo, totalParts int32) (err error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	dto := new(dto.FileStatus)
	r.db.Find(dto, fileID)
	if dto.FileID == 0 {
		return r.db.Create(dto).Error
	}
	return r.db.Save(dto).Error
}

func (r *repoFiles) SaveFileMessageMedia(fileID int64, req *msg.ClientSendMessageMedia) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	dto := new(dto.FileClientMessageMedia)
	r.db.Find(dto, fileID)
	if dto.FileID == 0 {
		dto.Map(fileID, req)
		return r.db.Create(dto).Error
	}
	return r.db.Save(dto).Error
}

func (r *repoFiles) GetFileMessageMedia(fileID int64) (*msg.ClientSendMessageMedia, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	dto := new(dto.FileClientMessageMedia)
	err := r.db.Find(dto, fileID).Error
	if err != nil {
		return nil, err
	}
	x := new(msg.ClientSendMessageMedia)
	dto.MapTo(x)
	return x, nil
}

func (r *repoFiles) GetDialogFileMessageMedia(peerID int64, peerType int32) ([]*msg.ClientSendMessageMedia, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	dtos := make([]dto.FileClientMessageMedia, 0)

	err := r.db.Where("PeerID=? AND PeerType=?", peerID, peerType).Find(&dtos).Error
	if err != nil {
		return nil, err
	}
	x := make([]*msg.ClientSendMessageMedia, 0, len(dtos))
	for _, d := range dtos {
		tmp := new(msg.ClientSendMessageMedia)
		d.MapTo(tmp)
		x = append(x, tmp)
	}

	return x, nil
}
