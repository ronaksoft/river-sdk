package repo

import (
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo/dto"
)

type RepoFiles interface {
	SaveFileStatus(fs *dto.FileStatus) (err error)
	GetAllFileStatus() []dto.FileStatus
	DeleteFileStatus(ID int64) error
	MoveUploadedFileToFiles(req *msg.ClientSendMessageMedia, sent *msg.MessagesSent) (err error)
	// delete this later
	GetFirstFileStatu() dto.FileStatus
}

type repoFiles struct {
	*repository
}

// SaveFileStatus
func (r *repoFiles) SaveFileStatus(fs *dto.FileStatus) (err error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	dto := new(dto.FileStatus)
	r.db.Find(dto, fs.MessageID)
	if dto.FileID == 0 {
		return r.db.Create(fs).Error
	}
	return r.db.Save(fs).Error
}

func (r *repoFiles) GetAllFileStatus() []dto.FileStatus {
	r.mx.Lock()
	defer r.mx.Unlock()

	dtos := make([]dto.FileStatus, 0)
	r.db.Find(&dtos)
	return dtos
}

// DeleteFileStatus
func (r *repoFiles) DeleteFileStatus(ID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	// remove pending message
	err := r.db.Where("MessageID = ?", ID).Delete(dto.FileStatus{}).Error

	return err
}

func (r *repoFiles) MoveUploadedFileToFiles(req *msg.ClientSendMessageMedia, sent *msg.MessagesSent) (err error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	f := new(dto.Files)
	r.db.Find(f, sent.MessageID)
	if f.MessageID > 0 {
		f.Map(sent.MessageID, sent.CreatedOn, req)
		err = r.db.Create(f).Error
	}
	f.Map(sent.MessageID, sent.CreatedOn, req)
	err = r.db.Save(f).Error
	return err
}

func (r *repoFiles) GetFirstFileStatu() dto.FileStatus {
	e := dto.FileStatus{}
	r.db.First(&e)
	return e
}
