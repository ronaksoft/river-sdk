package minirepo

import (
    "github.com/boltdb/bolt"
    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/rony/tools"
)

/*
   Creation Time: 2021 - May - 30
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

var (
    bucketTeams = []byte("TEAM")
)

type repoTeams struct {
    *repository
}

func newTeam(r *repository) *repoTeams {
    rd := &repoTeams{
        repository: r,
    }
    return rd
}

func (d *repoTeams) Save(teams ...*msg.Team) error {
    alloc := tools.NewAllocator()
    defer alloc.ReleaseAll()

    return d.db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket(bucketTeams)
        for _, team := range teams {
            err := b.Put(
                alloc.Gen(team.ID),
                alloc.Marshal(team),
            )
            if err != nil {
                return err
            }
        }
        return nil
    })
}

func (d *repoTeams) Delete(teamID int64) error {
    alloc := tools.NewAllocator()
    defer alloc.ReleaseAll()

    return d.db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket(bucketTeams)
        return b.Delete(alloc.Gen(teamID))
    })
}

func (d *repoTeams) Read(teamID int64) (*msg.Team, error) {
    alloc := tools.NewAllocator()
    defer alloc.ReleaseAll()

    team := &msg.Team{}
    err := d.db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket(bucketTeams)
        v := b.Get(alloc.Gen(teamID))
        if len(v) > 0 {
            return team.Unmarshal(v)
        }
        return domain.ErrNotFound
    })
    if err != nil {
        return nil, err
    }
    return team, nil
}

func (d *repoTeams) List() ([]*msg.Team, error) {
    teams := make([]*msg.Team, 0, 10)
    alloc := tools.NewAllocator()
    defer alloc.ReleaseAll()

    err := d.db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket(bucketTeams)
        return b.ForEach(func(k, v []byte) error {
            team := &msg.Team{}
            _ = team.Unmarshal(v)
            teams = append(teams, team)
            return nil
        })
    })

    if err != nil {
        return nil, err
    }
    return teams, nil
}
