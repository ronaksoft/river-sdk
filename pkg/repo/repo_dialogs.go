package repo

import (
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/msg/ext"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/dgraph-io/badger"
	"github.com/tidwall/buntdb"
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

func (r *repoDialogs) getDialogKey(peerID int64, peerType int32) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%021d.%d", prefixDialogs, peerID, peerType))
}

func (r *repoDialogs) getPinnedDialogKey(peerID int64, peerType int32) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%021d.%d", prefixPinnedDialogs, peerID, peerType))
}

func (r *repoDialogs) getPeerFromKey(key string) *msg.Peer {
	parts := strings.Split(key, ".")
	if len(parts) != 3 {
		return nil
	}
	return &msg.Peer{
		ID:   ronak.StrToInt64(parts[1]),
		Type: ronak.StrToInt32(parts[2]),
	}
}

func (r *repoDialogs) updateTopMessageID(peerID int64, peerType int32) {
	dialog := r.Get(peerID, peerType)
	if dialog == nil {
		return
	}
	_ = r.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = Messages.getPrefix(peerID, peerType)
		opts.Reverse = true
		it := txn.NewIterator(opts)
		it.Seek(Messages.getMessageKey(peerID, peerType, dialog.TopMessageID))
		if it.ValidForPrefix(opts.Prefix) {
			userMessage := new(msg.UserMessage)
			_ = it.Item().Value(func(val []byte) error {
				return userMessage.Unmarshal(val)
			})
			dialog.TopMessageID = userMessage.ID
			r.Save(dialog)
		}
		it.Close()
		return nil
	})
}

func (r *repoDialogs) updateLastUpdate(peerID int64, peerType int32, lastUpdate int64) {
	_ = r.bunt.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(
			ronak.ByteToStr(Dialogs.getDialogKey(peerID, peerType)),
			fmt.Sprintf("%021d", lastUpdate),
			nil,
		)

		return err
	})
}

func (r *repoDialogs) updateAccessHash(accessHash uint64, peerID int64, peerType int32) {
	dialog := r.Get(peerID, peerType)
	dialog.AccessHash = accessHash
	r.Save(dialog)
}

func (r *repoDialogs) countUnread(peerID int64, peerType int32, userID, maxID int64) (unread, mentioned int32) {
	_ = r.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = Messages.getPrefix(peerID, peerType)
		opts.Reverse = false
		it := txn.NewIterator(opts)
		for it.Seek(Messages.getMessageKey(peerID, peerType, maxID)); it.ValidForPrefix(opts.Prefix); it.Next() {
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
		return nil
	})
	return
}

func (r *repoDialogs) Get(peerID int64, peerType int32) *msg.Dialog {
	dialog := new(msg.Dialog)
	_ = r.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get(r.getDialogKey(peerID, peerType))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return dialog.Unmarshal(val)
		})
	})

	return dialog
}

func (r *repoDialogs) GetManyUsers(peerIDs []int64) []*msg.Dialog {
	dialogs := make([]*msg.Dialog, 0, len(peerIDs))
	for _, peerID := range peerIDs {
		dialog := r.Get(peerID, int32(msg.PeerUser))
		if dialog != nil {
			dialogs = append(dialogs, dialog)
		}
	}
	return dialogs
}

func (r *repoDialogs) GetManyGroups(peerIDs []int64) []*msg.Dialog {
	dialogs := make([]*msg.Dialog, 0, len(peerIDs))
	for _, peerID := range peerIDs {
		dialog := r.Get(peerID, int32(msg.PeerGroup))
		if dialog != nil {
			dialogs = append(dialogs, dialog)
		}
	}
	return dialogs
}

func (r *repoDialogs) SaveNew(dialog *msg.Dialog, lastUpdate int64) {
	r.Save(dialog)
	r.updateLastUpdate(dialog.PeerID, dialog.PeerType, lastUpdate)
}

func (r *repoDialogs) Save(dialog *msg.Dialog) {
	if dialog == nil {
		return
	}

	dialogBytes, _ := dialog.Marshal()
	_ = r.badger.Update(func(txn *badger.Txn) error {
		err := txn.SetEntry(badger.NewEntry(
			r.getDialogKey(dialog.PeerID, dialog.PeerType),
			dialogBytes,
		))
		if err != nil {
			return err
		}
		if dialog.Pinned {
			return txn.SetEntry(badger.NewEntry(
				r.getPinnedDialogKey(dialog.PeerID, dialog.PeerType),
				dialogBytes,
			))
		} else {
			return txn.Delete(r.getPinnedDialogKey(dialog.PeerID, dialog.PeerType))
		}
	})
}

func (r *repoDialogs) UpdateUnreadCount(peerID int64, peerType, unreadCount int32) {
	dialog := r.Get(peerID, peerType)
	if dialog == nil {
		return
	}

	dialog.UnreadCount = unreadCount
	r.Save(dialog)
	return
}

func (r *repoDialogs) UpdateReadInboxMaxID(userID, peerID int64, peerType int32, maxID int64) {
	dialog := r.Get(peerID, peerType)
	// current maxID is newer so skip updating dialog unread counts
	if dialog.ReadInboxMaxID > maxID || maxID > dialog.TopMessageID {
		return
	}
	dialog.ReadInboxMaxID = maxID
	dialog.UnreadCount, dialog.MentionedCount = r.countUnread(peerID, peerType, userID, maxID+1)
	r.Save(dialog)

	return
}

func (r *repoDialogs) UpdateReadOutboxMaxID(peerID int64, peerType int32, maxID int64) {
	dialog := r.Get(peerID, peerType)
	if maxID > dialog.TopMessageID {
		return
	}

	// current maxID is newer so skip updating dialog unread counts
	if dialog.ReadOutboxMaxID > maxID || maxID > dialog.TopMessageID {
		return
	}
	dialog.ReadOutboxMaxID = maxID
	r.Save(dialog)
	return
}

func (r *repoDialogs) UpdateNotifySetting(peerID int64, peerType int32, notifySettings *msg.PeerNotifySettings) {
	dialog := r.Get(peerID, peerType)
	dialog.NotifySettings = notifySettings
	r.Save(dialog)
	return
}

func (r *repoDialogs) UpdatePinned(in *msg.UpdateDialogPinned) {
	dialog := r.Get(in.Peer.ID, in.Peer.Type)
	dialog.Pinned = in.Pinned
	r.Save(dialog)
	return
}

func (r *repoDialogs) Delete(peerID int64, peerType int32) {
	_ = r.badger.Update(func(txn *badger.Txn) error {
		return txn.Delete(r.getDialogKey(peerID, peerType))
	})
}

func (r *repoDialogs) List(offset, limit int32) []*msg.Dialog {

	dialogs := make([]*msg.Dialog, 0, limit)
	_ = r.bunt.View(func(tx *buntdb.Tx) error {
		return tx.Descend(indexDialogs, func(key, value string) bool {
			if limit--; limit < 0 {
				return false
			}
			peer := r.getPeerFromKey(key)
			dialog := r.Get(peer.ID, peer.Type)
			if dialog != nil {
				dialogs = append(dialogs, dialog)
			}
			return true
		})
	})

	return dialogs
}

func (r *repoDialogs) GetPinnedDialogs() []*msg.Dialog {

	dialogs := make([]*msg.Dialog, 0, 7)
	_ = r.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = ronak.StrToByte(prefixDialogs)
		opts.Reverse = true
		it := txn.NewIterator(opts)
		for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
			dialog := new(msg.Dialog)
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
	return dialogs
}

func (r *repoDialogs) GetPeerIDs() []int64 {

	peerIDs := make([]int64, 0, 100)
	_ = r.bunt.View(func(tx *buntdb.Tx) error {
		return tx.Descend(indexDialogs, func(key, value string) bool {
			peer := r.getPeerFromKey(key)
			dialog := r.Get(peer.ID, peer.Type)
			if dialog != nil {
				peerIDs = append(peerIDs, peer.ID)
			}
			return true
		})
	})

	return peerIDs
}
