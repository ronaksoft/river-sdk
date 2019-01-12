package repo

import (
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo/dto"
)

type RepoFiles interface {
	SaveFileStatus(fs *dto.FileStatus) (err error)
	GetAllFileStatus() []dto.FileStatus
	GetFileStatus(msgID int64) (dto.FileStatus, error)
	DeleteFileStatus(ID int64) error
	MoveUploadedFileToFiles(req *msg.ClientSendMessageMedia, fileSize int32, sent *msg.MessagesSent) (err error)
	SaveFileDocument(msgID int64, doc *msg.MediaDocument) error
	GetExistingFileDocument(filePath string) *dto.Files
	GetFilePath(msgID, docID int64) string
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

func (r *repoFiles) GetFileStatus(msgID int64) (dto.FileStatus, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	mdl := dto.FileStatus{}
	err := r.db.Find(&mdl, msgID).Error
	return mdl, err
}

// DeleteFileStatus
func (r *repoFiles) DeleteFileStatus(ID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	// remove pending message
	err := r.db.Where("MessageID = ?", ID).Delete(dto.FileStatus{}).Error

	return err
}

func (r *repoFiles) MoveUploadedFileToFiles(req *msg.ClientSendMessageMedia, fileSize int32, sent *msg.MessagesSent) (err error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	f := new(dto.Files)
	r.db.Find(f, sent.MessageID)
	if f.MessageID > 0 {
		f.Map(sent.MessageID, sent.CreatedOn, fileSize, req)
		err = r.db.Create(f).Error
	}
	f.Map(sent.MessageID, sent.CreatedOn, fileSize, req)
	err = r.db.Save(f).Error
	return err
}

func (r *repoFiles) GetFirstFileStatu() dto.FileStatus {
	e := dto.FileStatus{}
	r.db.First(&e)
	return e
}

func (r *repoFiles) SaveFileDocument(msgID int64, doc *msg.MediaDocument) error {
	// 1. get file by documentID if we had file allready update file info for current msg too (use case : forwarded messages)
	existedDocument := dto.Files{}

	r.db.Table(existedDocument.TableName()).Where("DocumentID=?", doc.Doc.ID).First(&existedDocument)

	// 2. get file by messageID create or update document info
	mdl := dto.Files{}
	r.db.First(&mdl, msgID)
	if existedDocument.MessageID > 0 {
		mdl.MapFromFile(existedDocument)
	}

	if mdl.MessageID > 0 {
		mdl.MapFromDocument(doc)
		return r.db.Table(mdl.TableName()).Where("MessageID=?", msgID).Update(mdl).Error
	}
	mdl.MapFromDocument(doc)
	return r.db.Create(mdl).Error
}

func (r *repoFiles) GetExistingFileDocument(filePath string) *dto.Files {
	existedDocument := new(dto.Files)

	err := r.db.Table(existedDocument.TableName()).Where("FilePath=?", filePath).First(existedDocument).Error
	if err != nil {
		return nil
	}
	return existedDocument
}

func (r *repoFiles) GetFilePath(msgID, docID int64) string {
	f := dto.Files{}

	r.db.Find(&f, msgID)
	if f.MessageID > 0 && f.FilePath != "" {
		return f.FilePath
	}
	r.db.Table(f.TableName()).Where("DocumentID=?", docID).First(&f)
	if f.MessageID > 0 && f.FilePath != "" {
		return f.FilePath
	}
	return ""
}
