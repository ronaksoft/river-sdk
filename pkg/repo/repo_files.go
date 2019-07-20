package repo

import (
	"encoding/json"
	"fmt"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/dgraph-io/badger"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"
)

const (
	prefixFiles      = "FILE"
	prefixFileStatus = "FILE_STATUS"
)

type repoFiles struct {
	*repository
}

func (r *repoFiles) getKey(msgID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%012d", prefixFileStatus, msgID))
}

func (r *repoFiles) SaveStatus(fs *dto.FilesStatus) {
	r.mx.Lock()
	defer r.mx.Unlock()

	bytes, _ := json.Marshal(fs)
	_ = r.badger.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(r.getKey(fs.MessageID), bytes))
	})
}

func (r *repoFiles) GetAllStatuses() []dto.FilesStatus {
	r.mx.Lock()
	defer r.mx.Unlock()

	dtos := make([]dto.FilesStatus, 0)
	_ = r.badger.Update(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.Prefix = ronak.StrToByte(fmt.Sprintf("%s.", prefixFileStatus))
		it := txn.NewIterator(opt)
		for it.Rewind(); it.Valid(); it.Next() {
			_ = it.Item().Value(func(val []byte) error {
				fs := new(dto.FilesStatus)
				_ = json.Unmarshal(val, fs)
				dtos = append(dtos, *fs)
				return nil
			})
		}
		return nil
	})
	return dtos
}

func (r *repoFiles) GetStatus(msgID int64) (*dto.FilesStatus, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	mdl := new(dto.FilesStatus)
	err := r.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get(r.getKey(msgID))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, mdl)
		})
	})

	return mdl, err
}

func (r *repoFiles) DeleteStatus(msgID int64) {
	r.mx.Lock()
	defer r.mx.Unlock()

	_ = r.badger.Update(func(txn *badger.Txn) error {
		return txn.Delete(r.getKey(msgID))
	})
}

func (r *repoFiles) UpdateFileStatus(msgID int64, state domain.RequestStatus) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	fileStatus, err := r.GetStatus(msgID)
	if err != nil {
		return err
	}
	fileStatus.RequestStatus = int32(state)
	r.SaveStatus(fileStatus)
	return nil
}

func (r *repoFiles) MoveUploadedFileToFiles(req *msg.ClientSendMessageMedia, fileSize int32, sent *msg.MessagesSent)  {
	r.mx.Lock()
	defer r.mx.Unlock()

	fileStatus, err := r.GetStatus(sent.MessageID)
	if err != nil {
		fileStatus = new(dto.FilesStatus)
	}
	fileStatus.MessageID = sent.MessageID
	fileStatus.TotalSize = int64(fileSize)
	r.SaveStatus(fileStatus)
}

func (r *repoFiles) GetSharedMedia(peerID int64, peerType int32, mediaType int32) ([]*msg.UserMessage, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	limit := 50
	userMessages := make([]*msg.UserMessage, 0, limit)
	_ = r.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = Messages.getPrefix(peerID, peerType)
		opts.Reverse = true

		it := txn.NewIterator(opts)
		for it.Seek(Messages.getMessageKey(peerID, peerType, 1<<31)); it.Valid(); it.Next() {
			if limit--; limit < 0 {
				break
			}
			userMessage := new(msg.UserMessage)
			_ = it.Item().Value(func(val []byte) error {
				err := userMessage.Unmarshal(val)
				if err != nil {
					return err
				}
				if userMessage.MediaType == msg.MediaType(mediaType) {
					userMessages = append(userMessages, userMessage)
				}
				return nil
			})
		}
		it.Close()
		return nil
	})

	return userMessages, nil

}

// func (r *repoFiles) GetDBStatus() (map[int64]map[msg.DocumentAttributeType]dto.MediaInfo, error) {
// 	r.mx.Lock()
// 	defer r.mx.Unlock()
// 	var peerID int64
// 	var peerIDs []int64
//
// 	peerMediaSizeMap := make(map[int64]map[msg.DocumentAttributeType]dto.MediaInfo)
//
// 	f := dto.Files{}
//
// 	rows, err := r.db.Table(f.TableName()).Select("PeerID").Group("PeerID").Rows()
// 	defer rows.Close()
// 	if err != nil {
// 		logs.Warn(err.Error())
// 		return nil, err
// 	}
//
// 	for rows.Next() {
// 		err := rows.Scan(&peerID)
// 		if err != nil {
// 			logs.Warn("GetDBStatus", zap.String("rows::scan", err.Error()))
// 		} else {
// 			peerIDs = append(peerIDs, peerID)
// 		}
// 	}
// 	logs.Debug("peerIDs", zap.Any("", peerIDs))
//
// 	// range over peerIDs to collect each mediaType total size
// 	for _, peerId := range peerIDs {
// 		mMediaSize := make(map[msg.DocumentAttributeType]dto.MediaInfo)
// 		var f []dto.Files
// 		err := r.db.Where("PeerID = ? AND (FilePath <> ? OR ThumbFilePath <> ?)", peerId, "", "").Find(&f).Error
// 		if err != nil {
// 			logs.Warn(err.Error())
// 			return nil, err
// 		}
// 		if len(f) > 0 {
// 			for _, file := range f {
// 				attribs := make([]*msg.DocumentAttribute, 0)
// 				err = json.Unmarshal(file.Attributes, &attribs)
// 				if err != nil {
// 					logs.Warn(err.Error())
// 					return nil, err
// 				}
// 				for _, a := range attribs {
// 					switch a.Type {
// 					case msg.AttributeTypeFile:
// 						if len(attribs) > 1 {
// 							continue
// 						}
// 						if minfo, ok := mMediaSize[msg.AttributeTypeFile]; ok {
// 							minfo.MessageIDs = append(minfo.MessageIDs, file.MessageID)
// 							minfo.Size += file.FileSize
// 							mMediaSize[msg.AttributeTypeFile] = minfo
// 						} else {
// 							mediaInfo := dto.MediaInfo{
// 								MessageIDs: []int64{file.MessageID},
// 								Size:       file.FileSize,
// 							}
// 							mMediaSize[msg.AttributeTypeFile] = mediaInfo
// 						}
// 					case msg.AttributeTypeAudio:
// 						if minfo, ok := mMediaSize[msg.AttributeTypeAudio]; ok {
// 							minfo.MessageIDs = append(minfo.MessageIDs, file.MessageID)
// 							minfo.Size += file.FileSize
// 							mMediaSize[msg.AttributeTypeAudio] = minfo
// 						} else {
// 							mediaInfo := dto.MediaInfo{
// 								MessageIDs: []int64{file.MessageID},
// 								Size:       file.FileSize,
// 							}
// 							mMediaSize[msg.AttributeTypeAudio] = mediaInfo
// 						}
// 					case msg.AttributeTypeVideo:
// 						if minfo, ok := mMediaSize[msg.AttributeTypeVideo]; ok {
// 							minfo.MessageIDs = append(minfo.MessageIDs, file.MessageID)
// 							minfo.Size += file.FileSize
// 							mMediaSize[msg.AttributeTypeVideo] = minfo
// 						} else {
// 							mediaInfo := dto.MediaInfo{
// 								MessageIDs: []int64{file.MessageID},
// 								Size:       file.FileSize,
// 							}
// 							mMediaSize[msg.AttributeTypeVideo] = mediaInfo
// 						}
// 					case msg.AttributeTypePhoto:
// 						if minfo, ok := mMediaSize[msg.AttributeTypePhoto]; ok {
// 							minfo.MessageIDs = append(minfo.MessageIDs, file.MessageID)
// 							minfo.Size += file.FileSize
// 							mMediaSize[msg.AttributeTypePhoto] = minfo
// 						} else {
// 							mediaInfo := dto.MediaInfo{
// 								MessageIDs: []int64{file.MessageID},
// 								Size:       file.FileSize,
// 							}
// 							mMediaSize[msg.AttributeTypePhoto] = mediaInfo
// 						}
// 					case msg.AttributeAnimated:
// 						if minfo, ok := mMediaSize[msg.AttributeAnimated]; ok {
// 							minfo.MessageIDs = append(minfo.MessageIDs, file.MessageID)
// 							minfo.Size += file.FileSize
// 							mMediaSize[msg.AttributeAnimated] = minfo
// 						} else {
// 							mediaInfo := dto.MediaInfo{
// 								MessageIDs: []int64{file.MessageID},
// 								Size:       file.FileSize,
// 							}
// 							mMediaSize[msg.AttributeAnimated] = mediaInfo
// 						}
// 					case msg.AttributeTypeNone:
// 						if minfo, ok := mMediaSize[msg.AttributeTypeNone]; ok {
// 							minfo.MessageIDs = append(minfo.MessageIDs, file.MessageID)
// 							minfo.Size += file.FileSize
// 							mMediaSize[msg.AttributeTypeNone] = minfo
// 						} else {
// 							mediaInfo := dto.MediaInfo{
// 								MessageIDs: []int64{file.MessageID},
// 								Size:       file.FileSize,
// 							}
// 							mMediaSize[msg.AttributeTypeNone] = mediaInfo
// 						}
// 					default:
// 						// not implemented
// 					}
// 				}
// 			}
// 			peerMediaSizeMap[peerId] = mMediaSize
// 		}
// 	}
// 	return peerMediaSizeMap, nil
// }
//
// // ClearMedia returns media file paths to remove from device and updates database
// func (r *repoFiles) ClearMedia(messageIDs []int64) ([]string, error) {
// 	r.mx.Lock()
// 	defer r.mx.Unlock()
// 	dtoFiles := make([]dto.Files, 0, len(messageIDs))
// 	var filePaths []string
//
// 	f := dto.Files{}
// 	err := r.db.Table(f.TableName()).Where("MessageID in (?)", messageIDs).Find(&dtoFiles).Error
// 	if err != nil {
// 		logs.Warn(err.Error())
// 		return filePaths, err
// 	}
// 	for _, file := range dtoFiles {
// 		if file.FilePath != "" {
// 			filePaths = append(filePaths, file.FilePath)
// 		}
// 		if file.ThumbFilePath != "" {
// 			filePaths = append(filePaths, file.ThumbFilePath)
// 		}
// 	}
// 	err = r.db.Table(f.TableName()).Where("MessageID in (?)", messageIDs).Updates(map[string]interface{}{"FilePath": "", "ThumbFilePath": ""}).Error
// 	if err != nil {
// 		logs.Warn(err.Error())
// 		return filePaths, err
// 	}
// 	return filePaths, nil
// }
