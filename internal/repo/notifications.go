package repo

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/z"
	"github.com/dgraph-io/badger/v2"
	"github.com/ronaksoft/rony/pools"
	"github.com/ronaksoft/rony/tools"
)

/**
 * @created 25/05/2021 - 13:10
 * @project sdk
 * @author Reza Pilehvar
 */

const (
	prefixNotifications = "NOTIFICATIONS"
)

type repoNotifications struct {
	*repository
}

func (r *repoNotifications) getKey(teamID int64, peer *msg.InputPeer) []byte {
	sb := pools.AcquireStringsBuilder()
	sb.WriteString(prefixNotifications)
	sb.WriteRune('.')
	z.AppendStrInt64(sb, teamID)
	z.AppendStrInt64(sb, peer.ID)
	z.AppendStrInt32(sb, int32(peer.Type))
	sb.WriteRune('.')
	id := tools.StrToByte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func (r *repoNotifications) SetNotificationDismissTime(teamID int64, peer *msg.InputPeer, ts int64) error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		tsBytes := tools.StrToByte(tools.Int64ToStr(ts))
		key := r.getKey(teamID,peer)

		err := txn.SetEntry(badger.NewEntry(
			key, tsBytes,
		))

		return err
	})

	return err
}

func (r *repoNotifications) GetNotificationDismissTime(teamID int64, peer *msg.InputPeer) (int64, error) {
	var ts int64
	err := badgerView(func(txn *badger.Txn) error {
		item, err := txn.Get(r.getKey(teamID,peer))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			ts = tools.StrToInt64(tools.B2S(val))
			return nil
		})
		return err
	})

	return ts,err
}
