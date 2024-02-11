package contact

import (
    "bytes"
    "encoding/binary"
    "hash/crc32"
    "sort"

    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/river-sdk/internal/repo"
    "github.com/ronaksoft/river-sdk/internal/uiexec"
    "github.com/ronaksoft/rony"
    "go.uber.org/zap"
)

/*
   Creation Time: 2021 - May - 10
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func (r *contact) contactsImported(e *rony.MessageEnvelope) {
    x := new(msg.ContactsImported)
    if err := x.Unmarshal(e.Message); err != nil {
        r.Log().Error("couldn't unmarshal ContactsImported", zap.Error(err))
        return
    }

    r.Log().Debug("applies contactsImported")

    _ = repo.Users.SaveContact(domain.GetTeamID(e), x.ContactUsers...)
    _ = repo.Users.Save(x.Users...)
}

func (r *contact) contactsMany(e *rony.MessageEnvelope) {
    x := new(msg.ContactsMany)
    if err := x.Unmarshal(e.Message); err != nil {
        r.Log().Error("couldn't unmarshal ContactsMany", zap.Error(err))
        return
    }
    r.Log().Debug("applies contactsMany",
        zap.Int("Users", len(x.Users)),
        zap.Int("Contacts", len(x.Contacts)),
    )

    // If contacts are modified in server, then first clear all the contacts and rewrite the new ones
    if x.Modified {
        _ = repo.Users.DeleteAllContacts(domain.GetTeamID(e))
    }

    // Sort the contact users by their ids
    sort.Slice(x.ContactUsers, func(i, j int) bool { return x.ContactUsers[i].ID < x.ContactUsers[j].ID })

    _ = repo.Users.SaveContact(domain.GetTeamID(e), x.ContactUsers...)
    _ = repo.Users.Save(x.Users...)

    if len(x.ContactUsers) > 0 {
        buff := bytes.Buffer{}
        b := make([]byte, 8)
        for _, contactUser := range x.ContactUsers {
            binary.BigEndian.PutUint64(b, uint64(contactUser.ID))
            buff.Write(b)
        }
        crc32Hash := crc32.ChecksumIEEE(buff.Bytes())
        err := repo.System.SaveInt(domain.GetContactsGetHashKey(domain.GetTeamID(e)), uint64(crc32Hash))
        if err != nil {
            r.Log().Error("couldn't save ContactsHash in to the db", zap.Error(err))
        }
        uiexec.ExecDataSynced(false, true, false)
    }
}

func (r *contact) contactsTopPeers(e *rony.MessageEnvelope) {
    u := &msg.ContactsTopPeers{}
    err := u.Unmarshal(e.Message)
    if err != nil {
        r.Log().Error("couldn't unmarshal ContactsTopPeers", zap.Error(err))
        return
    }

    r.Log().Debug("applies ContactsTopPeers",
        zap.Int("L", len(u.Peers)),
        zap.String("Cat", u.Category.String()),
    )
    err = repo.TopPeers.Save(u.Category, r.SDK().SyncCtrl().GetUserID(), domain.GetTeamID(e), u.Peers...)
    if err != nil {
        r.Log().Error("got error on saving ContactsTopPeers", zap.Error(err))
    }
}
