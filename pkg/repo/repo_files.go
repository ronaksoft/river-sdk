package repo

import (
	"encoding/json"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"
	"go.uber.org/zap"
)

type repoFiles struct {
	*repository
}

func (r *repoFiles) SaveFileStatus(fs *dto.FilesStatus) (err error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	dto := new(dto.FilesStatus)
	r.db.Find(dto, fs.MessageID)
	if dto.FileID == 0 {
		return r.db.Create(fs).Error
	}
	return r.db.Save(fs).Error
}

func (r *repoFiles) GetAllFileStatus() []dto.FilesStatus {
	r.mx.Lock()
	defer r.mx.Unlock()

	dtos := make([]dto.FilesStatus, 0)
	r.db.Find(&dtos)
	return dtos
}

func (r *repoFiles) GetFileStatus(msgID int64) (*dto.FilesStatus, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	mdl := new(dto.FilesStatus)
	err := r.db.Find(mdl, msgID).Error
	return mdl, err
}

func (r *repoFiles) DeleteFileStatus(ID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	// remove pending status
	err := r.db.Where("MessageID = ?", ID).Delete(dto.FilesStatus{}).Error

	return err
}

func (r *repoFiles) DeleteManyFileStatus(msgIDs []int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	// remove pending status
	err := r.db.Where("MessageID IN (?)", msgIDs).Delete(dto.FilesStatus{}).Error

	return err
}

func (r *repoFiles) MoveUploadedFileToFiles(req *msg.ClientSendMessageMedia, fileSize int32, sent *msg.MessagesSent) (err error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	f := new(dto.Files)
	r.db.Find(f, sent.MessageID)
	if f.MessageID > 0 {
		f.Map(sent.MessageID, sent.CreatedOn, fileSize, req)
		return r.db.Table(f.TableName()).Where("MessageID=?", sent.MessageID).Update(f).Error
	}
	f.Map(sent.MessageID, sent.CreatedOn, fileSize, req)
	return r.db.Create(f).Error
}

func (r *repoFiles) SaveFileDocument(m *msg.UserMessage, doc *msg.MediaDocument) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	// 1. get file by documentID if we had file already update file info for current msg too (use case : forwarded messages)
	existedDocument := dto.Files{}
	// r.db.LogMode(true)
	// defer r.db.LogMode(false)

	// try to find any duplicated document
	r.db.Table(existedDocument.TableName()).Where("DocumentID=?", doc.Doc.ID).First(&existedDocument)

	// 2. get file by messageID create or update document info
	mdl := dto.Files{}
	r.db.First(&mdl, m.ID)
	fileExist := mdl.MessageID > 0

	// fill extra fields from existed document
	if existedDocument.MessageID > 0 {
		mdl.MapFromFile(existedDocument)
		mdl.MessageID = m.ID
	}
	// doc already exist
	if fileExist {
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

func (r *repoFiles) SaveDownloadingFile(fs *dto.FilesStatus) error {
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

	mdl := dto.FilesStatus{}
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

	r.mx.Lock()
	defer r.mx.Unlock()

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

func (r *repoFiles) UpdateFilePathByDocumentID(messageID int64, filePath string) {

	r.mx.Lock()
	defer r.mx.Unlock()

	m := dto.Messages{}
	r.db.Find(&m, messageID)
	if m.ID > 0 {
		switch msg.MediaType(m.MediaType) {
		case msg.MediaTypeEmpty:
			// NOP
		case msg.MediaTypePhoto:
			logs.Info("UpdateFilePathByDocumentID() Message.SharedMediaType is msg.MediaTypePhoto")
			// TODO:: implement it
		case msg.MediaTypeDocument:
			mediaDoc := new(msg.MediaDocument)
			err := mediaDoc.Unmarshal(m.Media)
			if err == nil {
				f := dto.Files{}
				r.db.Table(f.TableName()).Where("DocumentID=?", mediaDoc.Doc.ID).Updates(map[string]interface{}{
					"FilePath": filePath,
				})
			} else {
				logs.Error("UpdateFilePathByDocumentID()-> connat unmarshal MediaTypeDocument", zap.Error(err))
			}
		case msg.MediaTypeContact:
			logs.Info("UpdateFilePathByDocumentID() Message.SharedMediaType is msg.MediaTypeContact")
			// TODO:: implement it
		default:
			logs.Info("UpdateFilePathByDocumentID() Message.SharedMediaType is invalid")
		}
	}
}

func (r *repoFiles) GetFileByDocumentID(documentID int64) (*dto.Files, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	mdl := new(dto.Files)
	err := r.db.Where("DocumentID = ?", documentID).First(mdl).Error
	return mdl, err
}

func (r *repoFiles) GetDBStatus() (map[int64]map[msg.DocumentAttributeType]dto.MediaInfo, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	var peerID int64
	var peerIDs []int64

	peerMediaSizeMap := make(map[int64]map[msg.DocumentAttributeType]dto.MediaInfo)

	f := dto.Files{}

	rows, err := r.db.Table(f.TableName()).Select("PeerID").Group("PeerID").Rows()
	defer rows.Close()
	if err != nil {
		logs.Error(err.Error())
		return nil, err
	}

	for rows.Next() {
		err := rows.Scan(&peerID)
		if err != nil {
			logs.Error("GetDBStatus", zap.String("rows::scan", err.Error()))
		} else {
			peerIDs = append(peerIDs, peerID)
		}
	}
	logs.Debug("peerIDs", zap.Any("", peerIDs))

	// range over peerIDs to collect each mediaType total size
	for _, peerId := range peerIDs {
		mMediaSize := make(map[msg.DocumentAttributeType]dto.MediaInfo)
		var f []dto.Files
		err := r.db.Where("PeerID = ? AND (FilePath <> ? OR ThumbFilePath <> ?)", peerId, "", "").Find(&f).Error
		if err != nil {
			logs.Error(err.Error())
			return nil, err
		}
		if len(f) > 0 {
			for _, file := range f {
				attribs := make([]*msg.DocumentAttribute, 0)
				err = json.Unmarshal(file.Attributes, &attribs)
				if err != nil {
					logs.Error(err.Error())
					return nil, err
				}
				for _, a := range attribs {
					switch a.Type {
					case msg.AttributeTypeFile:
						if len(attribs) > 1 {
							continue
						}
						if minfo, ok := mMediaSize[msg.AttributeTypeFile]; ok {
							minfo.MessageIDs = append(minfo.MessageIDs, file.MessageID)
							minfo.Size += file.FileSize
							mMediaSize[msg.AttributeTypeFile] = minfo
						} else {
							mediaInfo := dto.MediaInfo{
								MessageIDs: []int64{file.MessageID},
								Size:       file.FileSize,
							}
							mMediaSize[msg.AttributeTypeFile] = mediaInfo
						}
					case msg.AttributeTypeAudio:
						if minfo, ok := mMediaSize[msg.AttributeTypeAudio]; ok {
							minfo.MessageIDs = append(minfo.MessageIDs, file.MessageID)
							minfo.Size += file.FileSize
							mMediaSize[msg.AttributeTypeAudio] = minfo
						} else {
							mediaInfo := dto.MediaInfo{
								MessageIDs: []int64{file.MessageID},
								Size:       file.FileSize,
							}
							mMediaSize[msg.AttributeTypeAudio] = mediaInfo
						}
					case msg.AttributeTypeVideo:
						if minfo, ok := mMediaSize[msg.AttributeTypeVideo]; ok {
							minfo.MessageIDs = append(minfo.MessageIDs, file.MessageID)
							minfo.Size += file.FileSize
							mMediaSize[msg.AttributeTypeVideo] = minfo
						} else {
							mediaInfo := dto.MediaInfo{
								MessageIDs: []int64{file.MessageID},
								Size:       file.FileSize,
							}
							mMediaSize[msg.AttributeTypeVideo] = mediaInfo
						}
					case msg.AttributeTypePhoto:
						if minfo, ok := mMediaSize[msg.AttributeTypePhoto]; ok {
							minfo.MessageIDs = append(minfo.MessageIDs, file.MessageID)
							minfo.Size += file.FileSize
							mMediaSize[msg.AttributeTypePhoto] = minfo
						} else {
							mediaInfo := dto.MediaInfo{
								MessageIDs: []int64{file.MessageID},
								Size:       file.FileSize,
							}
							mMediaSize[msg.AttributeTypePhoto] = mediaInfo
						}
					case msg.AttributeAnimated:
						if minfo, ok := mMediaSize[msg.AttributeAnimated]; ok {
							minfo.MessageIDs = append(minfo.MessageIDs, file.MessageID)
							minfo.Size += file.FileSize
							mMediaSize[msg.AttributeAnimated] = minfo
						} else {
							mediaInfo := dto.MediaInfo{
								MessageIDs: []int64{file.MessageID},
								Size:       file.FileSize,
							}
							mMediaSize[msg.AttributeAnimated] = mediaInfo
						}
					case msg.AttributeTypeNone:
						if minfo, ok := mMediaSize[msg.AttributeTypeNone]; ok {
							minfo.MessageIDs = append(minfo.MessageIDs, file.MessageID)
							minfo.Size += file.FileSize
							mMediaSize[msg.AttributeTypeNone] = minfo
						} else {
							mediaInfo := dto.MediaInfo{
								MessageIDs: []int64{file.MessageID},
								Size:       file.FileSize,
							}
							mMediaSize[msg.AttributeTypeNone] = mediaInfo
						}
					default:
						// not implemented
					}
				}
			}
		}
		peerMediaSizeMap[peerId] = mMediaSize
	}
	return peerMediaSizeMap, nil
}

// ClearMedia returns media file paths to remove from device and updates database
func (r *repoFiles) ClearMedia(messageIDs []int64) ([]string ,error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	dtoFiles := make([]dto.Files, 0, len(messageIDs))
	var filePaths []string

	f := dto.Files{}
	err := r.db.Table(f.TableName()).Where("MessageID in (?)", messageIDs).Find(&dtoFiles).Error
	if err != nil {
		logs.Error(err.Error())
		return filePaths, err
	}
	for _, file := range dtoFiles {
		if file.FilePath != "" {
			filePaths = append(filePaths, file.FilePath)
		}
		if file.ThumbFilePath != "" {
			filePaths = append(filePaths, file.ThumbFilePath)
		}
	}
	err = r.db.Table(f.TableName()).Where("MessageID in (?)", messageIDs).Updates(map[string]interface{}{"FilePath": "", "ThumbFilePath": ""}).Error
	if err != nil {
		logs.Error(err.Error())
		return filePaths, err
	}
	return filePaths, nil
}
