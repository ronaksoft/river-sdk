package mini

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/minirepo"
	"github.com/ronaksoft/rony"
	"go.uber.org/zap"
	"time"
)

/*
   Creation Time: 2021 - May - 07
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func (r *River) syncServerTime() (err error) {
	req := rony.PoolMessageEnvelope.Get()
	defer rony.PoolMessageEnvelope.Put(req)
	req.Fill(domain.NextRequestID(), msg.C_SystemGetServerTime, &msg.SystemGetServerTime{})

	r.network.HttpCommand(
		req,
		func() {
			err = domain.ErrRequestTimeout
		},
		func(m *rony.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_SystemServerTime:
				x := &msg.SystemServerTime{}
				err = x.Unmarshal(m.Message)
				if err != nil {
					logs.Error("SyncCtrl couldn't unmarshal SystemGetServerTime response", zap.Error(err))
					return
				}
				clientTime := time.Now().Unix()
				serverTime := x.Timestamp
				domain.TimeDelta = time.Duration(serverTime-clientTime) * time.Second

				logs.Debug("MiniRiver received SystemServerTime",
					zap.Int64("ServerTime", serverTime),
					zap.Int64("ClientTime", clientTime),
					zap.Duration("Difference", domain.TimeDelta),
				)
			case rony.C_Error:
				logs.Warn("MiniRiver received error on GetSystemServerTime", zap.Error(domain.ParseServerError(m.Message)))
				err = domain.ParseServerError(m.Message)
			}
		},
	)
	return
}

func (r *River) syncUpdateState() (updated bool, err error) {
	req := rony.PoolMessageEnvelope.Get()
	defer rony.PoolMessageEnvelope.Put(req)
	req.Fill(domain.NextRequestID(), msg.C_UpdateGetState, &msg.UpdateGetState{})

	currentUpdateID := r.getLastUpdateID()
	r.network.HttpCommand(
		req,
		func() {
			err = domain.ErrRequestTimeout
		},
		func(m *rony.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_UpdateState:
				x := &msg.UpdateState{}
				err = x.Unmarshal(m.Message)
				if err != nil {
					logs.Error("MiniRiver couldn't unmarshal SystemGetServerTime response", zap.Error(err))
					return
				}
				if x.UpdateID > currentUpdateID {
					updated = true
				}
				err = r.setLastUpdateID(x.UpdateID)
				if err != nil {
					logs.Error("MiniRiver couldn't save LastUpdateID to the database", zap.Error(err))
					return
				}
			case rony.C_Error:
				logs.Warn("MiniRiver received error on GetSystemServerTime", zap.Error(domain.ParseServerError(m.Message)))
				err = domain.ParseServerError(m.Message)
			}
		},
	)
	return
}

func (r *River) syncContacts() {
	req := rony.PoolMessageEnvelope.Get()
	defer rony.PoolMessageEnvelope.Put(req)
	req.Fill(domain.NextRequestID(), msg.C_ContactsGet, &msg.ContactsGet{Crc32Hash: r.getContactsHash()})
	r.network.HttpCommand(req,
		func() {

		},
		func(m *rony.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_ContactsMany:
				x := &msg.ContactsMany{}
				_ = x.Unmarshal(m.Message)
				err := minirepo.Users.SaveAllContacts(x)
				if err != nil {
					logs.Warn("MiniRiver got error on saving contacts", zap.Error(err))
					return
				}
				err = r.setContactsHash(x.Hash)
				if err != nil {
					logs.Warn("MiniRiver got error on saving contacts hash", zap.Error(err))
				}
			case rony.C_Error:
				x := &rony.Error{}
				_ = x.Unmarshal(m.Message)
				logs.Warn("MiniRiver got server error on syncing contacts", zap.Error(x))
			default:
				logs.Warn("MiniRiver got unknown server response")
			}
		},
	)
}

func (r *River) syncDialogs() {
	var (
		keepGoing       = true
		offset    int32 = 0
	)

	updated, err := r.syncUpdateState()
	if err != nil {
		logs.Warn("MiniRiver got error on UpdateSync", zap.Error(err))
	}
	if !updated {
		return
	}
	for keepGoing {
		req := rony.PoolMessageEnvelope.Get()
		req.Fill(domain.NextRequestID(), msg.C_MessagesGetDialogs, &msg.MessagesGetDialogs{
			Limit:         250,
			Offset:        offset,
			ExcludePinned: false,
		})
		r.network.HttpCommand(req,
			func() {

			},
			func(m *rony.MessageEnvelope) {
				switch m.Constructor {
				case msg.C_MessagesDialogs:
					x := &msg.MessagesDialogs{}
					_ = x.Unmarshal(m.Message)
					_ = minirepo.Dialogs.Save(x.Dialogs...)
					_ = minirepo.Users.SaveUser(x.Users...)
					_ = minirepo.Groups.Save(x.Groups...)
					offset += int32(len(x.Dialogs))
					if len(x.Dialogs) == 0 {
						keepGoing = false
					}
				case rony.C_Error:
					x := &rony.Error{}
					_ = x.Unmarshal(m.Message)
					logs.Warn("MiniRiver got server error on syncing dialogs", zap.Error(x))
				default:
					logs.Warn("MiniRiver got unknown server response")
				}
			},
		)
		rony.PoolMessageEnvelope.Put(req)
	}

}
