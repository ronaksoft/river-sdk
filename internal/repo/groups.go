package repo

import (
    "fmt"
    "strings"
    "time"

    "github.com/blevesearch/bleve/v2"
    "github.com/blevesearch/bleve/v2/search/query"
    "github.com/dgraph-io/badger/v2"
    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/river-sdk/internal/z"
    "github.com/ronaksoft/rony/pools"
    "github.com/ronaksoft/rony/tools"
)

const (
    prefixGroups             = "GRP"
    prefixGroupsFull         = "GRP_F"
    prefixGroupsParticipants = "GRP_P"
    prefixGroupsPhotoGallery = "GRP_PHG"
)

type repoGroups struct {
    *repository
}

func getGroupKey(groupID int64) []byte {
    sb := pools.AcquireStringsBuilder()
    sb.WriteString(prefixGroups)
    sb.WriteRune('.')
    z.AppendStrInt64(sb, groupID)
    id := tools.StrToByte(sb.String())
    pools.ReleaseStringsBuilder(sb)
    return id
}

func getGroupFullKey(groupID int64) []byte {
    sb := pools.AcquireStringsBuilder()
    sb.WriteString(prefixGroupsFull)
    sb.WriteRune('.')
    z.AppendStrInt64(sb, groupID)
    id := tools.StrToByte(sb.String())
    pools.ReleaseStringsBuilder(sb)
    return id
}

func getGroupByKey(txn *badger.Txn, groupKey []byte) (*msg.Group, error) {
    group := &msg.Group{}
    item, err := txn.Get(groupKey)
    if err != nil {
        return nil, err
    }
    err = item.Value(func(val []byte) error {
        return group.Unmarshal(val)
    })
    if err != nil {
        return nil, err
    }
    return group, nil
}

func getGroupFullByKey(txn *badger.Txn, groupFullKey []byte) (*msg.GroupFull, error) {
    groupFull := &msg.GroupFull{}
    item, err := txn.Get(groupFullKey)
    if err != nil {
        return nil, err
    }
    err = item.Value(func(val []byte) error {
        return groupFull.Unmarshal(val)
    })
    if err != nil {
        return nil, err
    }
    return groupFull, nil
}

func getGroupParticipantKey(groupID, memberID int64) []byte {
    sb := pools.AcquireStringsBuilder()
    sb.WriteString(prefixGroupsParticipants)
    sb.WriteRune('.')
    z.AppendStrInt64(sb, groupID)
    z.AppendStrInt64(sb, memberID)
    id := tools.StrToByte(sb.String())
    pools.ReleaseStringsBuilder(sb)
    return id
}

func getGroupPhotoGalleryKey(groupID, photoID int64) []byte {
    sb := pools.AcquireStringsBuilder()
    sb.WriteString(prefixGroupsPhotoGallery)
    sb.WriteRune('.')
    z.AppendStrInt64(sb, groupID)
    z.AppendStrInt64(sb, photoID)
    id := tools.StrToByte(sb.String())
    pools.ReleaseStringsBuilder(sb)
    return id
}

func getGroupPhotoGalleryPrefix(groupID int64) []byte {
    sb := pools.AcquireStringsBuilder()
    sb.WriteString(prefixGroupsPhotoGallery)
    sb.WriteRune('.')
    z.AppendStrInt64(sb, groupID)
    id := tools.StrToByte(sb.String())
    pools.ReleaseStringsBuilder(sb)
    return id
}

func saveGroup(txn *badger.Txn, group *msg.Group) error {
    groupKey := getGroupKey(group.ID)
    groupBytes, _ := group.Marshal()
    err := txn.SetEntry(badger.NewEntry(
        groupKey, groupBytes,
    ))
    if err != nil {
        return err
    }

    indexPeer(
        tools.ByteToStr(groupKey),
        GroupSearch{
            Type:   "group",
            Title:  group.Title,
            PeerID: group.ID,
            TeamID: fmt.Sprintf("%d", group.TeamID),
        },
    )

    err = saveGroupPhotos(txn, group.ID, group.Photo)
    if err != nil {
        return err
    }

    groupFull, _ := getGroupFullByKey(txn, getGroupFullKey(group.ID))
    if groupFull != nil {
        groupFull.Group = group
        err = saveGroupFull(txn, groupFull)
        if err != nil {
            return err
        }
    }

    return nil
}

func saveGroupFull(txn *badger.Txn, groupFull *msg.GroupFull) error {
    groupKey := getGroupFullKey(groupFull.Group.ID)
    groupBytes, _ := groupFull.Marshal()
    err := txn.SetEntry(badger.NewEntry(
        groupKey, groupBytes,
    ))
    if err != nil {
        return err
    }

    indexPeer(
        tools.ByteToStr(groupKey),
        GroupSearch{
            Type:   "group",
            Title:  groupFull.Group.Title,
            PeerID: groupFull.Group.ID,
            TeamID: fmt.Sprintf("%d", groupFull.Group.TeamID),
        },
    )

    err = saveGroupPhotos(txn, groupFull.Group.ID, groupFull.Group.Photo)
    if err != nil {
        return err
    }
    return nil
}

func removeGroupPhotoGallery(txn *badger.Txn, groupID int64, photoIDs ...int64) error {
    for _, photoID := range photoIDs {
        err := txn.Delete(getGroupPhotoGalleryKey(groupID, photoID))
        if err != nil && err != badger.ErrKeyNotFound {
            return err
        }
    }
    return nil
}

func (r *repoGroups) Save(groups ...*msg.Group) error {
    groupIDs := domain.MInt64B{}
    for _, v := range groups {
        groupIDs[v.ID] = true
    }

    return badgerUpdate(func(txn *badger.Txn) error {
        for _, group := range groups {
            err := saveGroup(txn, group)
            if err != nil {
                return err
            }
        }
        return nil
    })

}

func (r *repoGroups) SaveFull(group *msg.GroupFull) error {
    return badgerUpdate(func(txn *badger.Txn) error {
        return saveGroupFull(txn, group)
    })
}

func (r *repoGroups) GetMany(groupIDs []int64) ([]*msg.Group, error) {
    groups := make([]*msg.Group, 0, len(groupIDs))
    err := badgerView(func(txn *badger.Txn) error {
        for _, groupID := range groupIDs {
            if groupID == 0 {
                continue
            }
            group, err := getGroupByKey(txn, getGroupKey(groupID))
            switch err {
            case nil, badger.ErrKeyNotFound:
            default:
                return err
            }
            if group != nil {
                groups = append(groups, group)
            }
        }
        return nil
    })
    return groups, err
}

func (r *repoGroups) Get(groupID int64) (group *msg.Group, err error) {
    err = badgerView(func(txn *badger.Txn) error {
        group, err = getGroupByKey(txn, getGroupKey(groupID))
        if err != nil {
            return err
        }
        return nil
    })
    return
}

func (r *repoGroups) GetFull(groupID int64) (groupFull *msg.GroupFull, err error) {
    err = badgerView(func(txn *badger.Txn) error {
        groupFull, err = getGroupFullByKey(txn, getGroupFullKey(groupID))
        return err
    })
    return
}

func (r *repoGroups) AddParticipant(groupID int64, p *msg.GroupParticipant) error {
    return badgerUpdate(func(txn *badger.Txn) error {
        groupFull, err := getGroupFullByKey(txn, getGroupFullKey(groupID))
        if err != nil {
            return err
        }
        groupFull.Participants = append(groupFull.Participants, p)
        groupFull.Group.Participants = int32(len(groupFull.Participants))

        err = saveGroupFull(txn, groupFull)
        if err != nil {
            return err
        }

        return saveGroup(txn, groupFull.Group)
    })
}

func (r *repoGroups) RemoveParticipant(groupID int64, UserIDs ...int64) error {
    return badgerUpdate(func(txn *badger.Txn) error {
        groupFull, err := getGroupFullByKey(txn, getGroupFullKey(groupID))
        if err != nil {
            return err
        }
        pm := make(map[int64]*msg.GroupParticipant, len(groupFull.Participants))
        for _, p := range groupFull.Participants {
            pm[p.UserID] = p
        }
        for _, userID := range UserIDs {
            delete(pm, userID)
        }

        groupFull.Participants = groupFull.Participants[:0]
        for _, p := range pm {
            groupFull.Participants = append(groupFull.Participants, p)
        }
        groupFull.Group.Participants = int32(len(groupFull.Participants))
        err = saveGroupFull(txn, groupFull)
        if err != nil {
            return err
        }

        return saveGroup(txn, groupFull.Group)
    })
}

func (r *repoGroups) Delete(groupID int64) error {
    return badgerUpdate(func(txn *badger.Txn) error {
        err := txn.Delete(getGroupKey(groupID))
        switch err {
        case nil, badger.ErrKeyNotFound:
        default:
            return err
        }
        err = txn.Delete(getGroupFullKey(groupID))
        switch err {
        case nil, badger.ErrKeyNotFound:
        default:
            return err
        }
        return nil
    })
}

func (r *repoGroups) UpdatePhoto(groupID int64, groupPhoto *msg.GroupPhoto) error {
    return badgerUpdate(func(txn *badger.Txn) error {
        group, err := getGroupByKey(txn, getGroupKey(groupID))
        if err != nil {
            return err
        }
        group.Photo = groupPhoto
        return saveGroup(txn, group)
    })
}

func (r *repoGroups) RemovePhoto(groupID int64) error {
    return badgerUpdate(func(txn *badger.Txn) error {
        group, err := getGroupByKey(txn, getGroupKey(groupID))
        if err != nil {
            return err
        }
        group.Photo = nil

        err = removeGroupPhotoGallery(txn, groupID, group.Photo.PhotoID)
        if err != nil {
            return err
        }
        return saveGroup(txn, group)
    })
}

func (r *repoGroups) SavePhotoGallery(groupID int64, photos ...*msg.GroupPhoto) error {
    return badgerUpdate(func(txn *badger.Txn) error {
        return saveGroupPhotos(txn, groupID, photos...)
    })
}

func (r *repoGroups) RemovePhotoGallery(groupID int64, photoIDs ...int64) error {
    return badgerUpdate(func(txn *badger.Txn) error {
        return removeGroupPhotoGallery(txn, groupID, photoIDs...)
    })
}

func (r *repoGroups) GetPhotoGallery(groupID int64) ([]*msg.GroupPhoto, error) {
    photos := make([]*msg.GroupPhoto, 0, 5)
    err := badgerView(func(txn *badger.Txn) error {
        opts := badger.DefaultIteratorOptions
        opts.Prefix = getGroupPhotoGalleryPrefix(groupID)
        it := txn.NewIterator(opts)
        for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
            _ = it.Item().Value(func(val []byte) error {
                groupPhoto := new(msg.GroupPhoto)
                err := groupPhoto.Unmarshal(val)
                if err != nil {
                    return err
                }
                photos = append(photos, groupPhoto)
                return nil
            })
        }
        it.Close()
        return nil
    })
    return photos, err
}

func (r *repoGroups) UpdateTitle(groupID int64, title string) error {
    return badgerUpdate(func(txn *badger.Txn) error {
        group, err := getGroupByKey(txn, getGroupKey(groupID))
        if err != nil {
            return err
        }
        group.Title = title
        return saveGroup(txn, group)
    })
}

func (r *repoGroups) UpdateMemberType(groupID, userID int64, isAdmin bool) error {
    return badgerUpdate(func(txn *badger.Txn) error {
        group, err := getGroupByKey(txn, getGroupKey(groupID))
        if err != nil {
            return err
        }
        flags := make([]msg.GroupFlags, 0, len(group.Flags))
        for _, f := range group.Flags {
            if f != msg.GroupFlags_GroupFlagsAdmin {
                flags = append(flags, f)
            }
        }
        gp := new(msg.GroupParticipant)
        item, err := txn.Get(getGroupParticipantKey(groupID, userID))
        if err != nil {
            return err
        }
        err = item.Value(func(val []byte) error {
            return gp.Unmarshal(val)
        })
        if err != nil {
            return err
        }
        if isAdmin {
            flags = append(flags, msg.GroupFlags_GroupFlagsAdmin)
            gp.Type = msg.ParticipantType_ParticipantTypeAdmin
        } else {
            gp.Type = msg.ParticipantType_ParticipantTypeMember
        }
        group.Flags = flags
        groupParticipantKey := getGroupParticipantKey(groupID, gp.UserID)
        participantBytes, _ := gp.Marshal()
        err = txn.SetEntry(badger.NewEntry(
            groupParticipantKey, participantBytes,
        ))
        if err != nil {
            return err
        }

        return saveGroup(txn, group)
    })
}

func (r *repoGroups) ToggleAdmins(groupID int64, adminEnable bool) error {
    return badgerUpdate(func(txn *badger.Txn) error {
        group, err := getGroupByKey(txn, getGroupKey(groupID))
        if err != nil {
            return err
        }

        if adminEnable {
            for _, f := range group.Flags {
                if f == msg.GroupFlags_GroupFlagsAdminsEnabled {
                    return nil
                }
            }
            group.Flags = append(group.Flags, msg.GroupFlags_GroupFlagsAdminsEnabled)
        } else {
            for idx, f := range group.Flags {
                if f == msg.GroupFlags_GroupFlagsAdminsEnabled {
                    group.Flags[idx] = group.Flags[len(group.Flags)-1]
                    group.Flags = group.Flags[:len(group.Flags)-1]
                }
            }
        }

        return saveGroup(txn, group)
    })
}

func (r *repoGroups) Search(teamID int64, searchPhrase string) []*msg.Group {
    groups := make([]*msg.Group, 0, 100)
    if r.peerSearch == nil {
        return groups
    }

    t1 := bleve.NewTermQuery("group")
    t1.SetField("type")
    terms := strings.Fields(searchPhrase)
    qs := make([]query.Query, 0)
    for _, term := range terms {
        qs = append(qs, bleve.NewPrefixQuery(term), bleve.NewMatchQuery(term), bleve.NewFuzzyQuery(term))
    }
    t2 := bleve.NewDisjunctionQuery(qs...)
    t3 := bleve.NewTermQuery(fmt.Sprintf("%d", z.AbsInt64(teamID)))
    t3.SetField("team_id")
    searchRequest := bleve.NewSearchRequest(bleve.NewConjunctionQuery(t1, t2, t3))
    searchResult, _ := r.peerSearch.Search(searchRequest)
    _ = badgerView(func(txn *badger.Txn) error {
        for _, hit := range searchResult.Hits {
            group, _ := getGroupByKey(txn, tools.StrToByte(hit.ID))
            if group != nil {
                groups = append(groups, group)
            }
        }
        return nil
    })

    return groups
}

func (r *repoGroups) ReIndex() error {
    err := tools.Try(10, time.Second, func() error {
        if r.peerSearch == nil {
            return domain.ErrDoesNotExists
        }
        return nil
    })
    if err != nil {
        return err
    }
    return badgerView(func(txn *badger.Txn) error {
        opts := badger.DefaultIteratorOptions
        opts.Prefix = tools.StrToByte(prefixGroups)
        it := txn.NewIterator(opts)
        for it.Rewind(); it.Valid(); it.Next() {
            _ = it.Item().Value(func(val []byte) error {
                group := new(msg.Group)
                _ = group.Unmarshal(val)
                groupKey := tools.ByteToStr(getGroupKey(group.ID))
                if d, _ := r.peerSearch.Document(groupKey); d == nil {
                    indexPeer(
                        groupKey,
                        GroupSearch{
                            Type:   "group",
                            Title:  group.Title,
                            PeerID: group.ID,
                            TeamID: fmt.Sprintf("%d", group.TeamID),
                        },
                    )
                }
                return nil
            })
        }
        it.Close()
        return nil
    })

}

func (r *repoGroups) HasFlag(flags []msg.GroupFlags, flag msg.GroupFlags) bool {
    for _, f := range flags {
        if f == flag {
            return true
        }
    }
    return false
}
