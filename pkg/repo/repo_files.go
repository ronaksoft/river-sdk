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
	bytes, _ := json.Marshal(fs)
	_ = r.badger.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(r.getKey(fs.MessageID), bytes))
	})
}

func (r *repoFiles) GetAllStatuses() []dto.FilesStatus {
	dtos := make([]dto.FilesStatus, 0)
	_ = r.badger.Update(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.PrefetchValues = false
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
		it.Close()
		return nil
	})
	return dtos
}

func (r *repoFiles) GetStatus(msgID int64) (*dto.FilesStatus, error) {
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
	if err != nil {
		return nil, err
	}
	return mdl, nil
}

func (r *repoFiles) DeleteStatus(msgID int64) {
	_ = r.badger.Update(func(txn *badger.Txn) error {
		return txn.Delete(r.getKey(msgID))
	})
}

func (r *repoFiles) UpdateFileStatus(msgID int64, state domain.RequestStatus) {
	fileStatus, err := r.GetStatus(msgID)
	if err != nil {
		return
	}
	fileStatus.RequestStatus = int32(state)
	r.SaveStatus(fileStatus)
	return
}

func (r *repoFiles) GetSharedMedia(peerID int64, peerType int32, mediaType int32) ([]*msg.UserMessage, error) {
	limit := 50
	userMessages := make([]*msg.UserMessage, 0, limit)
	_ = r.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Prefix = Messages.getPrefix(peerID, peerType)
		opts.Reverse = true
		it := txn.NewIterator(opts)
		for it.Seek(Messages.getMessageKey(peerID, peerType, 1<<31)); it.ValidForPrefix(opts.Prefix); it.Next() {
			if limit--; limit < 0 {
				break
			}
			_ = it.Item().Value(func(val []byte) error {
				userMessage := new(msg.UserMessage)
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
