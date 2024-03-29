package auth

import (
    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/river-sdk/internal/repo"
    "github.com/ronaksoft/river-sdk/module"
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

type auth struct {
    module.Base
}

func New() *auth {
    r := &auth{}
    r.RegisterMessageAppliers(
        map[int64]domain.MessageApplier{
            msg.C_AuthAuthorization: r.authAuthorization,
            msg.C_AuthSentCode:      r.authSentCode,
        },
    )
    return r
}

func (r *auth) Name() string {
    return module.Auth
}

func (r *auth) authAuthorization(e *rony.MessageEnvelope) {
    x := new(msg.AuthAuthorization)
    if err := x.Unmarshal(e.Message); err != nil {
        r.Log().Error("AuthModule couldn't unmarshal AuthAuthorization", zap.Error(err))
        return
    }
    r.Log().Debug("AuthModule applies AuthAuthorization",
        zap.String("FirstName", x.User.FirstName),
        zap.String("LastName", x.User.LastName),
        zap.Int64("UserID", x.User.ID),
        zap.String("Bio", x.User.Bio),
        zap.String("Username", x.User.Username),
    )

    r.SDK().GetConnInfo().ChangeFirstName(x.User.FirstName)
    r.SDK().GetConnInfo().ChangeLastName(x.User.LastName)
    r.SDK().GetConnInfo().ChangeUserID(x.User.ID)
    r.SDK().GetConnInfo().ChangeBio(x.User.Bio)
    r.SDK().GetConnInfo().ChangeUsername(x.User.Username)
    if x.User.Phone != "" {
        r.SDK().GetConnInfo().ChangePhone(x.User.Phone)
    }
    r.SDK().GetConnInfo().Save()
    r.SDK().SyncCtrl().SetUserID(x.User.ID)

    repo.SetSelfUserID(x.User.ID)

    go func() {
        r.SDK().SyncCtrl().Sync()
    }()
}

func (r *auth) authSentCode(e *rony.MessageEnvelope) {
    x := new(msg.AuthSentCode)
    if err := x.Unmarshal(e.Message); err != nil {
        r.Log().Error("AuthModule couldn't unmarshal AuthSentCode", zap.Error(err))
        return
    }

    r.Log().Debug("AuthModule applies AuthSentCode")

    r.SDK().GetConnInfo().ChangePhone(x.Phone)
}
