package system

import (
    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/river-sdk/internal/repo"
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

func (r *system) systemConfig(e *rony.MessageEnvelope) {
    u := &msg.SystemConfig{}
    err := u.Unmarshal(e.Message)
    if err != nil {
        r.Log().Error("SystemModule couldn't unmarshal SystemConfig", zap.Error(err))
        return
    }

    r.Log().Debug("SystemModule applies SystemConfig")

    sysConfBytes, _ := u.Marshal()
    domain.SysConfig.Reactions = domain.SysConfig.Reactions[:0]
    err = domain.SysConfig.Unmarshal(sysConfBytes)
    if err != nil {
        r.Log().Error("SystemModule got error on unmarshalling SystemConfig", zap.Error(err))
    }
    err = repo.System.SaveBytes("SysConfig", sysConfBytes)
    if err != nil {
        r.Log().Error("SystemModule got error on saving SystemConfig", zap.Error(err))
    }
}
