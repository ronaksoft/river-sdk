package mini

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/minirepo"
	"git.ronaksoft.com/river/sdk/internal/request"
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
	r.network.HttpCommand(
		nil,
		request.NewCallback(
			0, 0, domain.NextRequestID(), msg.C_SystemGetServerTime, &msg.SystemGetServerTime{},
			func() {
				err = domain.ErrRequestTimeout
			},
			func(m *rony.MessageEnvelope) {
				switch m.Constructor {
				case msg.C_SystemServerTime:
					x := &msg.SystemServerTime{}
					err = x.Unmarshal(m.Message)
					if err != nil {
						logger.Error("couldn't unmarshal SystemGetServerTime response", zap.Error(err))
						return
					}
					clientTime := time.Now().Unix()
					serverTime := x.Timestamp
					domain.TimeDelta = time.Duration(serverTime-clientTime) * time.Second

					logger.Debug("MiniRiver received SystemServerTime",
						zap.Int64("ServerTime", serverTime),
						zap.Int64("ClientTime", clientTime),
						zap.Duration("Difference", domain.TimeDelta),
					)
				case rony.C_Error:
					logger.Warn("MiniRiver received error on GetSystemServerTime", zap.Error(domain.ParseServerError(m.Message)))
					err = domain.ParseServerError(m.Message)
				}
			},
			nil, false, 0, 0,
		),
	)
	return
}

func (r *River) syncUpdateState(teamID int64) (updated bool, err error) {
	req := rony.PoolMessageEnvelope.Get()
	defer rony.PoolMessageEnvelope.Put(req)
	req.Fill(domain.NextRequestID(), msg.C_UpdateGetState, &msg.UpdateGetState{})

	currentUpdateID := r.getLastUpdateID(teamID)
	r.network.HttpCommand(
		nil,
		request.NewCallback(0, 0, domain.NextRequestID(), msg.C_UpdateGetState, &msg.UpdateGetState{},
			func() {
				err = domain.ErrRequestTimeout
			},
			func(m *rony.MessageEnvelope) {
				switch m.Constructor {
				case msg.C_UpdateState:
					x := &msg.UpdateState{}
					err = x.Unmarshal(m.Message)
					if err != nil {
						logger.Error("MiniRiver couldn't unmarshal SystemGetServerTime response", zap.Error(err))
						return
					}
					if x.UpdateID > currentUpdateID {
						updated = true
					}
					err = r.setLastUpdateID(teamID, x.UpdateID)
					if err != nil {
						logger.Error("MiniRiver couldn't save LastUpdateID to the database", zap.Error(err))
						return
					}
				case rony.C_Error:
					logger.Warn("MiniRiver received error on GetSystemServerTime", zap.Error(domain.ParseServerError(m.Message)))
					err = domain.ParseServerError(m.Message)
				}
			},
			nil, false, 0, 0,
		),
	)
	return
}

func (r *River) syncContacts(teamID int64, teamAccess uint64) {
	r.network.HttpCommand(
		nil,
		request.NewCallback(
			teamID, teamAccess, domain.NextRequestID(), msg.C_ContactsGet, &msg.ContactsGet{Crc32Hash: r.getContactsHash(teamID)},
			func() {},
			func(m *rony.MessageEnvelope) {
				switch m.Constructor {
				case msg.C_ContactsMany:
					x := &msg.ContactsMany{}
					_ = x.Unmarshal(m.Message)
					if !x.Modified {
						return
					}
					err := minirepo.Users.SaveAllContacts(teamID, x)
					if err != nil {
						logger.Warn("MiniRiver got error on saving users/contacts", zap.Error(err))
						return
					}
					err = r.setContactsHash(teamID, x.Hash)
					if err != nil {
						logger.Warn("MiniRiver got error on saving contacts hash", zap.Error(err))
					}
					r.mainDelegate.DataSynced(false, true, false)
				case rony.C_Error:
					x := &rony.Error{}
					_ = x.Unmarshal(m.Message)
					logger.Warn("MiniRiver got server error on syncing contacts", zap.Error(x))
				default:
					logger.Warn("MiniRiver got unknown server response")
				}
			},
			nil, false, 0, 0,
		),
	)

}

func (r *River) syncDialogs(teamID int64, teamAccess uint64) {
	var (
		keepGoing       = true
		offset    int32 = 0
	)

	updated, err := r.syncUpdateState(teamID)
	if err != nil {
		logger.Warn("MiniRiver got error on UpdateSync", zap.Error(err))
	}
	if !updated {
		return
	}
	for keepGoing {
		r.network.HttpCommand(
			nil,
			request.NewCallback(
				teamID, teamAccess, domain.NextRequestID(), msg.C_MessagesGetDialogs,
				&msg.MessagesGetDialogs{
					Limit:         250,
					Offset:        offset,
					ExcludePinned: false,
				},
				func() {},
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
						logger.Warn("MiniRiver got server error on syncing dialogs", zap.Error(x))
					default:
						logger.Warn("MiniRiver got unknown server response")
					}
				},
				nil, false, 0, domain.HttpRequestTimeShort,
			),
		)
	}

	r.mainDelegate.DataSynced(true, false, false)
}

func (r *River) syncTeams() {
	r.network.HttpCommand(
		nil,
		request.NewCallback(
			0, 0, domain.NextRequestID(), msg.C_AccountGetTeams, &msg.AccountGetTeams{},
			func() {},
			func(m *rony.MessageEnvelope) {
				switch m.Constructor {
				case msg.C_TeamsMany:
					x := &msg.TeamsMany{}
					_ = x.Unmarshal(m.Message)

					err := minirepo.Teams.Save(x.Teams...)
					if err != nil {
						logger.Warn("got error on saving teams [Teams]", zap.Error(err))
						return
					}
					err = minirepo.Users.SaveUser(x.Users...)
					if err != nil {
						logger.Warn("got error on saving teams [Users]", zap.Error(err))
						return
					}
				case rony.C_Error:
					x := &rony.Error{}
					_ = x.Unmarshal(m.Message)
					logger.Warn("MiniRiver got server error on syncing contacts", zap.Error(x))
				default:
					logger.Warn("MiniRiver got unknown server response")
				}
			},
			nil, false, 0, 0,
		),
	)

}
