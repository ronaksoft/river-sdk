package repo

import (
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/z"
	"github.com/dgraph-io/badger/v2"
	"github.com/ronaksoft/rony/pools"
	"github.com/ronaksoft/rony/tools"
)

/**
 * @created 02/09/2020 - 11:57
 * @project riversdk
 * @author reza
 */

type repoTeams struct {
	*repository
}

const (
	prefixTeams = "TEAMS"
)

func getTeamKey(teamID int64) []byte {
	sb := pools.AcquireStringsBuilder()
	sb.WriteString(prefixTeams)
	sb.WriteRune('.')
	z.AppendStrInt64(sb, teamID)
	id := tools.StrToByte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func (r *repoTeams) List() []*msg.Team {
	teamList := make([]*msg.Team, 0, 10)
	err := badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = tools.StrToByte(fmt.Sprintf("%s.", prefixTeams))
		it := txn.NewIterator(opts)

		defer it.Close()
		for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
			_ = it.Item().Value(func(val []byte) error {
				t := &msg.Team{}
				err := t.Unmarshal(val)
				if err != nil {
					return err
				}
				teamList = append(teamList, t)
				return nil
			})
		}

		return nil
	})
	logs.ErrorOnErr("RepoTeams got error on GetAll", err)
	return teamList
}

func (r *repoTeams) Get(teamID int64) (team *msg.Team, err error) {
	err = badgerView(func(txn *badger.Txn) error {
		team, err = r.get(txn, teamID)
		return err
	})
	return
}

func (r *repoTeams) get(txn *badger.Txn, teamID int64) (*msg.Team, error) {
	team := &msg.Team{}
	item, err := txn.Get(getTeamKey(teamID))
	if err != nil {
		return nil, err
	}
	err = item.Value(func(val []byte) error {
		return team.Unmarshal(val)
	})
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (r *repoTeams) Save(teams ...*msg.Team) error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		for _, team := range teams {
			teamBytes, _ := team.Marshal()
			recentSearchKey := getTeamKey(team.ID)
			err := txn.SetEntry(badger.NewEntry(
				recentSearchKey, teamBytes,
			))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (r *repoTeams) Delete(teamID int64) error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		teamKey := getTeamKey(teamID)
		err := txn.Delete(teamKey)
		return err
	})
	return err
}

func (r *repoTeams) Clear() error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = tools.StrToByte(fmt.Sprintf("%s.", prefixTeams))
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
			err := txn.Delete(it.Item().KeyCopy(nil))
			if err != nil {
				return err
			}
		}
		it.Close()

		return nil
	})
	return err
}
