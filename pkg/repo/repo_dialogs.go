package repo

import (
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/dgraph-io/badger"
	"github.com/tidwall/buntdb"
	"go.uber.org/zap"
	"strings"
)

const (
	prefixDialogs       = "DLG"
	prefixPinnedDialogs = "PDLG"

	indexDialogs = prefixDialogs
)

type repoDialogs struct {
	*repository
}

func getDialogKey(peerID int64, peerType int32) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%021d.%d", prefixDialogs, peerID, peerType))
}

func getPinnedDialogKey(peerID int64, peerType int32) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%021d.%d", prefixPinnedDialogs, peerID, peerType))
}

func getPeerFromKey(key string) *msg.Peer {
	parts := strings.Split(key, ".")
	if len(parts) != 3 {
		return nil
	}
	return &msg.Peer{
		ID:   ronak.StrToInt64(parts[1]),
		Type: ronak.StrToInt32(parts[2]),
	}
}

func updateTopMessageID(txn *badger.Txn, dialog *msg.Dialog) error {
	var topMessageID int64
	opts := badger.DefaultIteratorOptions
	opts.Prefix = getMessagePrefix(dialog.PeerID, dialog.PeerType)
	opts.Reverse = true
	it := txn.NewIterator(opts)
	it.Seek(getMessageKey(dialog.PeerID, dialog.PeerType, dialog.TopMessageID))
	if it.ValidForPrefix(opts.Prefix) {
		userMessage := new(msg.UserMessage)
		_ = it.Item().Value(func(val []byte) error {
			return userMessage.Unmarshal(val)
		})
		topMessageID = userMessage.ID
	}
	it.Close()

	dialog.TopMessageID = topMessageID
	return saveDialog(txn, dialog)
}

func saveDialog(txn *badger.Txn, dialog *msg.Dialog) error {
	dialogBytes, _ := dialog.Marshal()
	err := txn.SetEntry(badger.NewEntry(
		getDialogKey(dialog.PeerID, dialog.PeerType),
		dialogBytes,
	))
	if err != nil {
		return err
	}
	if dialog.Pinned {
		return txn.SetEntry(badger.NewEntry(
			getPinnedDialogKey(dialog.PeerID, dialog.PeerType),
			dialogBytes,
		))
	} else {
		return txn.Delete(getPinnedDialogKey(dialog.PeerID, dialog.PeerType))
	}
}

func getDialog(txn *badger.Txn, peerID int64, peerType int32) (*msg.Dialog, error) {
	dialog := &msg.Dialog{}
	item, err := txn.Get(getDialogKey(peerID, peerType))
	if err != nil {
		return nil, err
	}
	err = item.Value(func(val []byte) error {
		return dialog.Unmarshal(val)
	})
	if err != nil {
		return nil, err
	}
	return dialog, nil
}

func updateDialogAccessHash(txn *badger.Txn, accessHash uint64, peerID int64, peerType int32) error {
	dialog, err := getDialog(txn, peerID, peerType)
	if err != nil {
		return err
	}
	dialog.AccessHash = accessHash
	return saveDialog(txn, dialog)
}

func countDialogUnread(txn *badger.Txn, peerID int64, peerType int32, userID, maxID int64) (unread, mentioned int32, err error) {
	opts := badger.DefaultIteratorOptions
	opts.Prefix = getMessagePrefix(peerID, peerType)
	opts.Reverse = false
	it := txn.NewIterator(opts)
	for it.Seek(getMessageKey(peerID, peerType, maxID)); it.ValidForPrefix(opts.Prefix); it.Next() {
		_ = it.Item().Value(func(val []byte) error {
			userMessage := new(msg.UserMessage)
			_ = userMessage.Unmarshal(val)
			if userMessage.SenderID != userID {
				unread++
			}
			for _, entity := range userMessage.Entities {
				if entity.Type == msg.MessageEntityTypeMention && entity.UserID == userID {
					mentioned++
				}
			}
			return nil
		})
	}
	it.Close()
	return
}

func (r *repoDialogs) updateLastUpdate(peerID int64, peerType int32, lastUpdate int64) {
	_ = r.bunt.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(
			ronak.ByteToStr(getDialogKey(peerID, peerType)),
			fmt.Sprintf("%021d", lastUpdate),
			nil,
		)
		return err
	})
}

func (r *repoDialogs) Get(peerID int64, peerType int32) (dialog *msg.Dialog, err error) {
	err = r.badger.View(func(txn *badger.Txn) error {
		dialog, err = getDialog(txn, peerID, peerType)
		return err
	})
	return
}

func (r *repoDialogs) SaveNew(dialog *msg.Dialog, lastUpdate int64) (err error) {
	return r.badger.Update(func(txn *badger.Txn) error {
		err = saveDialog(txn, dialog)
		if err != nil {
			return err
		}
		r.updateLastUpdate(dialog.PeerID, dialog.PeerType, lastUpdate)
		return nil
	})
}

func (r *repoDialogs) Save(dialog *msg.Dialog) {
	if dialog == nil {
		logs.Error("RepoDialog calls save for nil dialog")
		return
	}
	err := r.badger.Update(func(txn *badger.Txn) error {
		return saveDialog(txn, dialog)
	})
	logs.ErrorOnErr("RepoDialog got error on save dialog", err)
}

func (r *repoDialogs) UpdateUnreadCount(peerID int64, peerType, unreadCount int32) {
	err := r.badger.Update(func(txn *badger.Txn) error {
		dialog, err := getDialog(txn, peerID, peerType)
		if err != nil {
			return err
		}
		dialog.UnreadCount = unreadCount
		return saveDialog(txn, dialog)
	})
	logs.ErrorOnErr("RepoDialog got error on update unread counter", err,
		zap.Int64("PeerID", peerID),
		zap.Int32("PeerType", peerType),
	)
}

func (r *repoDialogs) UpdateReadInboxMaxID(userID, peerID int64, peerType int32, maxID int64) {
	err := r.badger.Update(func(txn *badger.Txn) error {
		dialog, err := getDialog(txn, peerID, peerType)
		if err != nil {
			return err
		}
		// current maxID is newer so skip updating dialog unread counts
		if dialog.ReadInboxMaxID > maxID || maxID > dialog.TopMessageID {
			return nil
		}
		dialog.ReadInboxMaxID = maxID
		dialog.UnreadCount, dialog.MentionedCount, err = countDialogUnread(txn, peerID, peerType, userID, maxID+1)
		return saveDialog(txn, dialog)
	})
	logs.ErrorOnErr("RepoDialog got error on update read inbox maxID", err,
		zap.Int64("UserID", userID),
		zap.Int64("PeerID", peerID),
		zap.Int32("PeerType", peerType),
		zap.Int64("MaxID", maxID),
	)
	return
}

func (r *repoDialogs) UpdateReadOutboxMaxID(peerID int64, peerType int32, maxID int64) {
	err := r.badger.Update(func(txn *badger.Txn) error {
		dialog, err := getDialog(txn, peerID, peerType)
		if err != nil {
			return err
		}
		// current maxID is newer so skip updating dialog unread counts
		if dialog.ReadOutboxMaxID > maxID || maxID > dialog.TopMessageID {
			return nil
		}
		dialog.ReadOutboxMaxID = maxID
		return saveDialog(txn, dialog)
	})
	logs.ErrorOnErr("RepoDialog got error on update read outbox maxID", err,
		zap.Int64("MaxID", maxID),
		zap.Int64("PeerID", peerID),
		zap.Int32("PeerType", peerType),
	)
	return
}

func (r *repoDialogs) UpdateNotifySetting(peerID int64, peerType int32, notifySettings *msg.PeerNotifySettings) {
	err := r.badger.Update(func(txn *badger.Txn) error {
		dialog, err := getDialog(txn, peerID, peerType)
		if err != nil {
			return err
		}
		dialog.NotifySettings = notifySettings
		return saveDialog(txn, dialog)
	})
	logs.ErrorOnErr("RepoDialog got error on update notify setting", err,
		zap.Int64("PeerID", peerID),
		zap.Int32("PeerType", peerType),
	)
	return
}

func (r *repoDialogs) UpdatePinned(in *msg.UpdateDialogPinned) {
	err := r.badger.Update(func(txn *badger.Txn) error {
		dialog, err := getDialog(txn, in.Peer.ID, in.Peer.Type)
		if err != nil {
			return err
		}
		dialog.Pinned = in.Pinned
		return saveDialog(txn, dialog)
	})
	logs.ErrorOnErr("RepoDialog got error on update pin", err,
		zap.Int64("PeerID", in.Peer.ID),
		zap.Int32("PeerType", in.Peer.Type),
	)
	return
}

func (r *repoDialogs) Delete(peerID int64, peerType int32) {
	err := r.badger.Update(func(txn *badger.Txn) error {
		return txn.Delete(getDialogKey(peerID, peerType))
	})
	logs.Error("RepoDialogs got error on deleting dialog",
		zap.Error(err),
		zap.Int64("PeerID", peerID),
		zap.Int32("PeerType", peerType),
	)
}

func (r *repoDialogs) List(offset, limit int32) []*msg.Dialog {
	dialogs := make([]*msg.Dialog, 0, limit)
	err := r.badger.View(func(txn *badger.Txn) error {
		return r.bunt.View(func(tx *buntdb.Tx) error {
			return tx.Descend(indexDialogs, func(key, value string) bool {
				if limit--; limit < 0 {
					return false
				}
				peer := getPeerFromKey(key)
				dialog, err := getDialog(txn, peer.ID, peer.Type)
				if err == nil && dialog != nil {
					dialogs = append(dialogs, dialog)
				}
				return true
			})
		})
	})

	logs.ErrorOnErr("RepoDialogs got error on getting list", err)
	return dialogs
}

func (r *repoDialogs) GetPinnedDialogs() []*msg.Dialog {
	dialogs := make([]*msg.Dialog, 0, 7)
	err := r.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = ronak.StrToByte(prefixDialogs)
		opts.Reverse = true
		it := txn.NewIterator(opts)
		for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
			dialog := &msg.Dialog{}
			_ = it.Item().Value(func(val []byte) error {
				err := dialog.Unmarshal(val)
				if err != nil {
					return err
				}
				if dialog.Pinned {
					dialogs = append(dialogs, dialog)
				}
				return nil
			})
		}
		it.Close()
		return nil
	})
	logs.Error("RepoDialogs got error on getting pinned dialogs", zap.Error(err))
	return dialogs
}

func (r *repoDialogs) GetPeerIDs() []int64 {
	peerIDs := make([]int64, 0, 100)
	err := r.badger.View(func(txn *badger.Txn) error {
		return r.bunt.View(func(tx *buntdb.Tx) error {
			return tx.Descend(indexDialogs, func(key, value string) bool {
				peer := getPeerFromKey(key)
				dialog, err := getDialog(txn, peer.ID, peer.Type)
				if err == nil && dialog != nil {
					peerIDs = append(peerIDs, peer.ID)
				}
				return true
			})
		})
	})
	logs.WarnOnErr("RepoDialogs got error on get peer ids", err)
	return peerIDs
}
