package repo

import (
	"fmt"
	"git.ronaksoft.com/river/msg/msg"
	"git.ronaksoft.com/ronak/riversdk/internal/logs"
	"git.ronaksoft.com/ronak/riversdk/pkg/domain"
	"github.com/dgraph-io/badger/v2"
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
	return domain.StrToByte(fmt.Sprintf("%s.%021d", prefixTeams, teamID))
}

func (r *repoTeams) List() []*msg.Team {
	teamList := make([]*msg.Team, 0, 100)

	err := badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = domain.StrToByte(fmt.Sprintf("%s.", prefixTeams))
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

func (r *repoTeams) Put(team *msg.Team) error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		return r.PutWithTransaction(txn, team)
	})
	return err
}

func (r *repoTeams) PutMany(teams ...*msg.Team) error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		for _, team := range teams {
			err := r.PutWithTransaction(txn, team)

			if err != nil {
				return err
			}
		}

		return nil
	})
	return err
}

func (r *repoTeams) PutWithTransaction(txn *badger.Txn, team *msg.Team) error {
	teamBytes, _ := team.Marshal()
	recentSearchKey := getTeamKey(team.ID)
	err := txn.SetEntry(badger.NewEntry(
		recentSearchKey, teamBytes,
	))

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
		opts.Prefix = domain.StrToByte(fmt.Sprintf("%s.", prefixTeams))
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
