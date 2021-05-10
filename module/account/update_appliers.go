package account

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/repo"
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

func (r *account) updateAccountPrivacy(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateAccountPrivacy)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Debug("SyncCtrl applies UpdateAccountPrivacy",
		zap.Int64("UpdateID", x.UpdateID),
	)

	err = repo.Account.SetPrivacy(msg.PrivacyKey_PrivacyKeyChatInvite, x.ChatInvite)
	logs.WarnOnErr("SyncCtrl got error on set privacy (ChatInvite)", err)
	err = repo.Account.SetPrivacy(msg.PrivacyKey_PrivacyKeyLastSeen, x.LastSeen)
	logs.WarnOnErr("SyncCtrl got error on set privacy (LastSeen)", err)
	err = repo.Account.SetPrivacy(msg.PrivacyKey_PrivacyKeyPhoneNumber, x.PhoneNumber)
	logs.WarnOnErr("SyncCtrl got error on set privacy (PhoneNumber)", err)
	err = repo.Account.SetPrivacy(msg.PrivacyKey_PrivacyKeyProfilePhoto, x.ProfilePhoto)
	logs.WarnOnErr("SyncCtrl got error on set privacy (ProfilePhoto)", err)
	err = repo.Account.SetPrivacy(msg.PrivacyKey_PrivacyKeyForwardedMessage, x.ForwardedMessage)
	logs.WarnOnErr("SyncCtrl got error on set privacy (ForwardedMessage)", err)
	err = repo.Account.SetPrivacy(msg.PrivacyKey_PrivacyKeyCall, x.Call)
	logs.WarnOnErr("SyncCtrl got error on set privacy (Call)", err)
	res := []*msg.UpdateEnvelope{u}
	return res, nil
}
