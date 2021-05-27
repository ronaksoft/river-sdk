package repo

import (
	"context"
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/z"
	"github.com/dgraph-io/badger/v2"
	"github.com/ronaksoft/rony/pools"
	"github.com/ronaksoft/rony/tools"
	"github.com/tidwall/buntdb"
	"strings"
	"sync/atomic"
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
	z.AppendStrInt64(sb, teamID)
	z.AppendStrInt64(sb, peerID)
	z.AppendStrInt32(sb, peerType)
	id := []byte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func getDialogPrefix(teamID int64) []byte {
	sb := pools.AcquireStringsBuilder()
	sb.WriteString(prefixDialogs)
	sb.WriteRune('.')
	z.AppendStrInt64(sb, teamID)
	id := []byte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func getPinnedDialogKey(teamID int64, peerID int64, peerType int32) []byte {
	sb := pools.AcquireStringsBuilder()
	sb.WriteString(prefixPinnedDialogs)
	sb.WriteRune('.')
	z.AppendStrInt64(sb, teamID)
	z.AppendStrInt64(sb, peerID)
	z.AppendStrInt32(sb, peerType)
	id := tools.StrToByte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func getDialogPeerFromIndexKey(key string) (int64, *msg.Peer) {
	parts := strings.Split(key, ".")
	if len(parts) != 4 {
		return 0, nil
	}
	return tools.StrToInt64(parts[1]), &msg.Peer{
		ID:   tools.StrToInt64(parts[2]),
		Type: tools.StrToInt32(parts[3]),
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
	return getDialogByKey(txn, getDialogKey(teamID, peerID, peerType))
}

func getDialogByKey(txn *badger.Txn, key []byte) (*msg.Dialog, error) {
	dialog := &msg.Dialog{}
	item, err := txn.Get(key)
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

func countDialogUnread(txn *badger.Txn, teamID, peerID int64, peerType int32, userID, maxID int64) (unread, mentioned int32, err error) {
	opts := badger.DefaultIteratorOptions
	opts.Prefix = getMessagePrefix(teamID, peerID, peerType)
	opts.Reverse = false
	it := txn.NewIterator(opts)
	for it.Seek(getMessageKey(teamID, peerID, peerType, maxID)); it.ValidForPrefix(opts.Prefix); it.Next() {
		_ = it.Item().Value(func(val []byte) error {
			userMessage := &msg.UserMessage{}
			_ = userMessage.Unmarshal(val)
			if userMessage.SenderID != userID {
				unread++
			}
			for _, entity := range userMessage.Entities {
				switch {
				case entity.Type == msg.MessageEntityType_MessageEntityTypeMention && entity.UserID == userID:
					fallthrough
				case entity.Type == msg.MessageEntityType_MessageEntityTypeMentionAll && userMessage.SenderID != userID:
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

func (r *repoDialogs) SaveNew(dialog *msg.Dialog, lastUpdate int64) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		err := saveDialog(txn, dialog)
		if err != nil {
			return err
		}
		return updateDialogLastUpdate(dialog.TeamID, dialog.PeerID, dialog.PeerType, lastUpdate)
	})
}

func (r *repoDialogs) Save(dialog *msg.Dialog) error {
	if dialog == nil {
		return nil
	}
	return badgerUpdate(func(txn *badger.Txn) error {
		err := saveDialog(txn, dialog)
		if err != nil {
			return err
		}
		return nil
	})
}

func (r *repoDialogs) UpdateReadInboxMaxID(userID, teamID, peerID int64, peerType int32, maxID int64) error {
	return badgerUpdate(func(txn *badger.Txn) error {
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
		if err != nil {
			return err
		}
		return saveDialog(txn, dialog)
	})
}

func (r *repoDialogs) UpdateReadOutboxMaxID(teamID, peerID int64, peerType int32, maxID int64) error {
	return badgerUpdate(func(txn *badger.Txn) error {
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
}

func (r *repoDialogs) UpdateNotifySetting(teamID, peerID int64, peerType int32, notifySettings *msg.PeerNotifySettings) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		dialog, err := getDialog(txn, teamID, peerID, peerType)
		if err != nil {
			return err
		}
		dialog.NotifySettings = notifySettings
		return saveDialog(txn, dialog)
	})
}

func (r *repoDialogs) UpdatePinned(in *msg.UpdateDialogPinned) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		dialog, err := getDialog(txn, in.TeamID, in.Peer.ID, in.Peer.Type)
		if err != nil {
			return err
		}
		dialog.Pinned = in.Pinned
		return saveDialog(txn, dialog)
	})
}

func (r *repoDialogs) UpdateCallStarted(in *msg.UpdatePhoneCallStarted) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		dialog, err := getDialog(txn, in.TeamID, in.Peer.ID, in.Peer.Type)
		if err != nil {
			return err
		}
		dialog.ActiveCallID = in.CallId
		return saveDialog(txn, dialog)
	})
}

func (r *repoDialogs) UpdateCallEnded(in *msg.UpdatePhoneCallEnded) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		dialog, err := getDialog(txn, in.TeamID, in.Peer.ID, in.Peer.Type)
		if err != nil {
			return err
		}
		dialog.ActiveCallID = 0
		return saveDialog(txn, dialog)
	})
}

func (r *repoDialogs) UpdatePinMessageID(teamID int64, peerID int64, peerType int32, messageID int64) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		dialog, err := getDialog(txn, teamID, peerID, peerType)
		if err != nil {
			return err
		}
		dialog.PinnedMessageID = messageID
		return saveDialog(txn, dialog)
	})
}

func (r *repoDialogs) Delete(teamID, peerID int64, peerType int32) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		return txn.Delete(getDialogKey(teamID, peerID, peerType))
	})
}

func (r *repoDialogs) List(teamID int64, offset, limit int32) ([]*msg.Dialog, error) {
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
				tID, peer := getDialogPeerFromIndexKey(key)
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
	if err != nil {
		return nil, err
	}
	return dialogs, nil
}

func (r *repoDialogs) CountDialogs(teamID int64) int32 {
	var cnt int32
	st := r.badger.NewStream()
	st.Prefix = getDialogPrefix(teamID)
	st.ChooseKey = func(item *badger.Item) bool {
		atomic.AddInt32(&cnt, 1)
		return false
	}
	_ = st.Orchestrate(context.Background())
	return cnt
}

func (r *repoDialogs) GetPinnedDialogs() []*msg.Dialog {
	dialogs := make([]*msg.Dialog, 0, 7)
	_ = badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = tools.StrToByte(prefixDialogs)
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

	return dialogs
}

func (r *repoDialogs) CountAllUnread(userID, teamID int64, mutes bool) (unread, mentioned int32, err error) {
	err = badgerView(func(txn *badger.Txn) error {
		st := r.badger.NewStream()
		st.Prefix = getDialogPrefix(teamID)
		st.ChooseKey = func(item *badger.Item) bool {
			d, err := getDialogByKey(txn, item.Key())
			if err != nil {
				return false
			}
			u, m, err := countDialogUnread(txn, d.TeamID, d.PeerID, d.PeerType, userID, d.ReadInboxMaxID+1)
			if err != nil {
				return false
			}
			if mutes || (d.NotifySettings != nil && d.NotifySettings.MuteUntil < domain.Now().Unix()) {
				atomic.AddInt32(&unread, u)
			}
			atomic.AddInt32(&mentioned, m)
			return false
		}
		st.Send = func(list *badger.KVList) error {
			return nil
		}
		return st.Orchestrate(context.Background())
	})

	return
}
