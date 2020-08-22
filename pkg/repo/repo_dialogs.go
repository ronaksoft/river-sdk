package repo

import (
	"fmt"
	"git.ronaksoft.com/river/msg/msg"
	"git.ronaksoft.com/ronak/riversdk/internal/logs"
	"git.ronaksoft.com/ronak/riversdk/internal/pools"
	"git.ronaksoft.com/ronak/riversdk/internal/tools"
	"git.ronaksoft.com/ronak/riversdk/pkg/domain"
	"github.com/dgraph-io/badger/v2"
	"github.com/tidwall/buntdb"
	"go.uber.org/zap"
	"strings"
)

const (
	prefixDialogs       = "DLG"
	prefixPinnedDialogs = "PDLG"
	indexDialogs        = prefixDialogs
)

type repoDialogs struct {
	*repository
}

func getDialogKey(teamID int64, peerID int64, peerType int32) []byte {
	sb := pools.AcquireStringsBuilder()
	sb.WriteString(prefixDialogs)
	sb.WriteRune('.')
	tools.AppendStrInt64(sb, teamID)
	tools.AppendStrInt64(sb, peerID)
	tools.AppendStrInt32(sb, peerType)
	id := tools.StrToByte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func getPinnedDialogKey(teamID int64, peerID int64, peerType int32) []byte {
	sb := pools.AcquireStringsBuilder()
	sb.WriteString(prefixPinnedDialogs)
	sb.WriteRune('.')
	tools.AppendStrInt64(sb, teamID)
	tools.AppendStrInt64(sb, peerID)
	tools.AppendStrInt32(sb, peerType)
	id := tools.StrToByte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func getPeerFromIndexKey(key string) (int64, *msg.Peer) {
	parts := strings.Split(key, ".")
	if len(parts) != 4 {
		return 0, nil
	}
	return tools.StrToInt64(parts[1]), &msg.Peer{
		ID:   domain.StrToInt64(parts[2]),
		Type: domain.StrToInt32(parts[3]),
	}
}

func saveDialog(txn *badger.Txn, dialog *msg.Dialog) error {
	dialogBytes, _ := dialog.Marshal()
	err := txn.SetEntry(badger.NewEntry(
		getDialogKey(dialog.TeamID, dialog.PeerID, dialog.PeerType),
		dialogBytes,
	))
	if err != nil {
		return err
	}
	if dialog.Pinned {
		return txn.SetEntry(badger.NewEntry(
			getPinnedDialogKey(dialog.TeamID, dialog.PeerID, dialog.PeerType),
			dialogBytes,
		))
	} else {
		return txn.Delete(getPinnedDialogKey(dialog.TeamID, dialog.PeerID, dialog.PeerType))
	}
}

func getDialog(txn *badger.Txn, teamID, peerID int64, peerType int32) (*msg.Dialog, error) {
	dialog := &msg.Dialog{}
	item, err := txn.Get(getDialogKey(teamID, peerID, peerType))
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

func updateDialogAccessHash(txn *badger.Txn, accessHash uint64, teamID, peerID int64, peerType int32) error {
	dialog, err := getDialog(txn, teamID, peerID, peerType)
	if err != nil {
		return err
	}
	dialog.AccessHash = accessHash
	return saveDialog(txn, dialog)
}

func countDialogUnread(txn *badger.Txn, teamID, peerID int64, peerType int32, userID, maxID int64) (unread, mentioned int32, err error) {
	opts := badger.DefaultIteratorOptions
	opts.Prefix = getMessagePrefix(teamID, peerID, peerType)
	opts.Reverse = false
	it := txn.NewIterator(opts)
	for it.Seek(getMessageKey(teamID, peerID, peerType, maxID)); it.ValidForPrefix(opts.Prefix); it.Next() {
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

func updateDialogLastUpdate(teamID int64, peerID int64, peerType int32, lastUpdate int64) error {
	return r.bunt.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(
			fmt.Sprintf("%s.%d.%d.%d", indexDialogs, teamID, peerID, peerType),
			fmt.Sprintf("%021d", lastUpdate),
			nil,
		)
		return err
	})
}

func (r *repoDialogs) Get(teamID, peerID int64, peerType int32) (dialog *msg.Dialog, err error) {
	err = badgerView(func(txn *badger.Txn) error {
		dialog, err = getDialog(txn, teamID, peerID, peerType)
		return err
	})
	return
}

func (r *repoDialogs) SaveNew(dialog *msg.Dialog, lastUpdate int64) (err error) {
	return badgerUpdate(func(txn *badger.Txn) error {
		err = saveDialog(txn, dialog)
		if err != nil {
			return err
		}
		updateDialogLastUpdate(dialog.TeamID, dialog.PeerID, dialog.PeerType, lastUpdate)
		return nil
	})
}

func (r *repoDialogs) Save(dialog *msg.Dialog) error {
	if dialog == nil {
		logs.Error("RepoDialog calls save for nil dialog")
		return nil
	}
	err := badgerUpdate(func(txn *badger.Txn) error {
		err := saveDialog(txn, dialog)
		if err != nil {
			return err
		}
		return nil
	})
	logs.ErrorOnErr("RepoDialog got error on save dialog", err)
	return err
}

func (r *repoDialogs) UpdateUnreadCount(teamID, peerID int64, peerType, unreadCount int32) {
	err := badgerUpdate(func(txn *badger.Txn) error {
		dialog, err := getDialog(txn, teamID, peerID, peerType)
		if err != nil {
			return err
		}
		dialog.UnreadCount = unreadCount
		return saveDialog(txn, dialog)
	})
	logs.ErrorOnErr("RepoDialog got error on update unread counter", err,
		zap.Int64("TeamID", teamID),
		zap.Int64("PeerID", peerID),
		zap.Int32("PeerType", peerType),
	)
}

func (r *repoDialogs) UpdateReadInboxMaxID(userID, teamID, peerID int64, peerType int32, maxID int64) {
	err := badgerUpdate(func(txn *badger.Txn) error {
		dialog, err := getDialog(txn, teamID, peerID, peerType)
		if err != nil {
			return err
		}
		// current maxID is newer so skip updating dialog unread counts
		if dialog.ReadInboxMaxID > maxID || maxID > dialog.TopMessageID {
			return nil
		}
		dialog.ReadInboxMaxID = maxID
		dialog.UnreadCount, dialog.MentionedCount, err = countDialogUnread(txn, teamID, peerID, peerType, userID, maxID+1)
		return saveDialog(txn, dialog)
	})
	logs.ErrorOnErr("RepoDialog got error on update read inbox maxID", err,
		zap.Int64("UserID", userID),
		zap.Int64("TeamID", teamID),
		zap.Int64("PeerID", peerID),
		zap.Int32("PeerType", peerType),
		zap.Int64("MaxID", maxID),
	)
	return
}

func (r *repoDialogs) UpdateReadOutboxMaxID(teamID, peerID int64, peerType int32, maxID int64) {
	err := badgerUpdate(func(txn *badger.Txn) error {
		dialog, err := getDialog(txn, teamID, peerID, peerType)
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
		zap.Int64("TeamID", teamID),
		zap.Int64("PeerID", peerID),
		zap.Int32("PeerType", peerType),
	)
	return
}

func (r *repoDialogs) UpdateNotifySetting(teamID, peerID int64, peerType int32, notifySettings *msg.PeerNotifySettings) {
	err := badgerUpdate(func(txn *badger.Txn) error {
		dialog, err := getDialog(txn, teamID, peerID, peerType)
		if err != nil {
			return err
		}
		dialog.NotifySettings = notifySettings
		return saveDialog(txn, dialog)
	})
	logs.ErrorOnErr("RepoDialog got error on update notify setting", err,
		zap.Int64("TeamID", teamID),
		zap.Int64("PeerID", peerID),
		zap.Int32("PeerType", peerType),
	)
	return
}

func (r *repoDialogs) UpdatePinned(in *msg.UpdateDialogPinned) {
	err := badgerUpdate(func(txn *badger.Txn) error {
		dialog, err := getDialog(txn, in.TeamID, in.Peer.ID, in.Peer.Type)
		if err != nil {
			return err
		}
		dialog.Pinned = in.Pinned
		return saveDialog(txn, dialog)
	})
	logs.ErrorOnErr("RepoDialog got error on update pin", err,
		zap.Int64("TeamID", in.TeamID),
		zap.Int64("PeerID", in.Peer.ID),
		zap.Int32("PeerType", in.Peer.Type),
	)
	return
}

func (r *repoDialogs) Delete(teamID, peerID int64, peerType int32) {
	err := badgerUpdate(func(txn *badger.Txn) error {
		return txn.Delete(getDialogKey(teamID, peerID, peerType))
	})
	if err != nil {
		logs.Error("RepoDialogs got error on deleting dialog",
			zap.Error(err),
			zap.Int64("TeamID", teamID),
			zap.Int64("PeerID", peerID),
			zap.Int32("PeerType", peerType),
		)
	}

}

func (r *repoDialogs) List(teamID int64, offset, limit int32) []*msg.Dialog {
	dialogs := make([]*msg.Dialog, 0, limit)
	err := badgerView(func(txn *badger.Txn) error {
		return r.bunt.View(func(tx *buntdb.Tx) error {
			return tx.Descend(indexDialogs, func(key, value string) bool {
				if offset--; offset >= 0 {
					return true
				}
				if limit--; limit < 0 {
					return false
				}
				tID, peer := getPeerFromIndexKey(key)
				if tID != teamID {
					return true
				}
				dialog, err := getDialog(txn, teamID, peer.ID, peer.Type)
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
	err := badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = domain.StrToByte(prefixDialogs)
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
	logs.ErrorOnErr("RepoDialogs got error on getting pinned dialogs", err)
	return dialogs
}
