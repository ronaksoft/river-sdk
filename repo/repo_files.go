package repo

import (
	"encoding/json"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo/dto"
)

// Files repoFiles interface
type Files interface {
	SaveFileStatus(fs *dto.FileStatus) (err error)
	GetAllFileStatus() []dto.FileStatus
	GetFileStatus(msgID int64) (*dto.FileStatus, error)
	DeleteFileStatus(msgID int64) error
	DeleteManyFileStatus(msgIDs []int64) error
	MoveUploadedFileToFiles(req *msg.ClientSendMessageMedia, fileSize int32, sent *msg.MessagesSent) (err error)
	SaveFileDocument(m *msg.UserMessage, doc *msg.MediaDocument) error
	GetExistingFileDocument(filePath string) *dto.Files
	GetFilePath(msgID, docID int64) string
	SaveDownloadingFile(fs *dto.FileStatus) error
	UpdateFileStatus(msgID int64, state domain.RequestStatus) error
	UpdateThumbnailPath(msgID int64, filePath string) error
	GetFile(msgID int64) (*dto.Files, error)

	// delete this later
	GetFirstFileStatu() dto.FileStatus

	GetSharedMedia(peerID int64, peerType int32, mediaType int32) ([]*msg.UserMessage, error)
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

func (r *repoFiles) GetFileStatus(msgID int64) (*dto.FileStatus, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	mdl := new(dto.FileStatus)
	err := r.db.Find(mdl, msgID).Error
	return mdl, err
}

// DeleteFileStatus
func (r *repoFiles) DeleteFileStatus(ID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	// remove pending status
	err := r.db.Where("MessageID = ?", ID).Delete(dto.FileStatus{}).Error

	return err
}

// DeleteManyFileStatus
func (r *repoFiles) DeleteManyFileStatus(msgIDs []int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	// remove pending status
	err := r.db.Where("MessageID IN (?)", msgIDs).Delete(dto.FileStatus{}).Error

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
	r.mx.Lock()
	defer r.mx.Unlock()

	e := dto.FileStatus{}
	r.db.First(&e)
	return e
}

func (r *repoFiles) SaveFileDocument(m *msg.UserMessage, doc *msg.MediaDocument) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	// 1. get file by documentID if we had file allready update file info for current msg too (use case : forwarded messages)
	existedDocument := dto.Files{}
	// r.db.LogMode(true)
	// defer r.db.LogMode(false)

	r.db.Table(existedDocument.TableName()).Where("DocumentID=?", doc.Doc.ID).First(&existedDocument)

	// 2. get file by messageID create or update document info
	mdl := dto.Files{}
	r.db.First(&mdl, m.ID)
	if existedDocument.MessageID > 0 {
		mdl.MapFromFile(existedDocument)
		mdl.MessageID = m.ID
	}
	// doc already exist
	if mdl.MessageID > 0 {
		mdl.MapFromDocument(doc)
		return r.db.Table(mdl.TableName()).Where("MessageID=?", m.ID).Update(&mdl).Error
	}

	mdl.MapFromUserMessageDocument(m, doc)
	return r.db.Create(&mdl).Error

}

func (r *repoFiles) GetExistingFileDocument(filePath string) *dto.Files {
	r.mx.Lock()
	defer r.mx.Unlock()

	existedDocument := new(dto.Files)

	err := r.db.Table(existedDocument.TableName()).Where("FilePath=?", filePath).First(existedDocument).Error
	if err != nil {
		return nil
	}
	return existedDocument
}

func (r *repoFiles) GetFilePath(msgID, docID int64) string {
	r.mx.Lock()
	defer r.mx.Unlock()

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

func (r *repoFiles) SaveDownloadingFile(fs *dto.FileStatus) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	mdl := dto.Files{}
	r.db.Find(&mdl, fs.MessageID)
	if mdl.MessageID == 0 {
		(&mdl).MapFromFileStatus(fs)
		r.db.Create(&mdl)
	}
	return r.db.Table(mdl.TableName()).Where("MessageID=? OR DocumentID=?", fs.MessageID, fs.FileID).Updates(map[string]interface{}{
		"FilePath": fs.FilePath,
	}).Error
}

func (r *repoFiles) UpdateFileStatus(msgID int64, state domain.RequestStatus) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	mdl := dto.FileStatus{}
	return r.db.Table(mdl.TableName()).Where("MessageID=? ", msgID).Updates(map[string]interface{}{
		"RequestStatus": int32(state),
	}).Error
}

func (r *repoFiles) UpdateThumbnailPath(msgID int64, filePath string) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	e := new(dto.Files)
	return r.db.Table(e.TableName()).Where("MessageID = ? ", msgID).Updates(map[string]interface{}{
		"ThumbFilePath": filePath,
	}).Error

}

func (r *repoFiles) GetFile(msgID int64) (*dto.Files, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	mdl := new(dto.Files)
	err := r.db.Find(mdl, msgID).Error
	return mdl, err
}

func (r *repoFiles) GetSharedMedia(peerID int64, peerType int32, mediaType int32) ([]*msg.UserMessage, error) {
	mType := domain.SharedMediaType(mediaType)

	dtos := make([]dto.Files, 0)
	err := r.db.Where("PeerID=? AND PeerType=?", peerID, peerType).Find(&dtos).Error
	if err != nil {
		return nil, err
	}

	// extract msgIDs by applying attribute filter
	msgIDs := domain.MInt64B{}
	for _, d := range dtos {
		attribs := make([]*msg.DocumentAttribute, 0)
		err := json.Unmarshal(d.Attributes, &attribs)
		if err != nil {
			continue
		}
	attributes:
		for _, a := range attribs {
			// AttributeTypeNone
			// AttributeTypeAudio
			// AttributeTypeVideo
			// AttributeTypePhoto
			// AttributeTypeFile
			// AttributeAnimated
			switch mType {
			case domain.SharedMediaTypeAll:
				msgIDs[d.MessageID] = true
				break attributes
			case domain.SharedMediaTypeFile:
				if a.Type == msg.AttributeTypeNone || a.Type == msg.AttributeTypeFile {
					msgIDs[d.MessageID] = true
				}
			case domain.SharedMediaTypeMedia:
				if a.Type == msg.AttributeTypePhoto || a.Type == msg.AttributeAnimated || a.Type == msg.AttributeTypeVideo {
					msgIDs[d.MessageID] = true
				}
			case domain.SharedMediaTypeVoice:
				if a.Type == msg.AttributeTypeAudio {
					x := new(msg.DocumentAttributeAudio)
					err := x.Unmarshal(a.Data)
					if err == nil {
						if x.Voice {
							msgIDs[d.MessageID] = true
						}
					}
				}
			case domain.SharedMediaTypeAudio:
				if a.Type == msg.AttributeTypeAudio {
					x := new(msg.DocumentAttributeAudio)
					err := x.Unmarshal(a.Data)
					if err == nil {
						if !x.Voice {
							msgIDs[d.MessageID] = true
						}
					}
				}
			case domain.SharedMediaTypeLink:
				// not implemented
			default:
				// not implemented
			}
		}
	}

	messageIDs := msgIDs.ToArray()
	messages := make([]*msg.UserMessage, 0, len(messageIDs))
	dtoMsgs := make([]dto.Messages, 0, len(messageIDs))
	err = r.db.Where("ID in (?)", messageIDs).Find(&dtoMsgs).Error
	if err != nil {
		return nil, err
	}

	for _, v := range dtoMsgs {

		tmp := new(msg.UserMessage)
		v.MapTo(tmp)
		messages = append(messages, tmp)
	}

	return messages, nil

}
