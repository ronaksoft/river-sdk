package account

import (
	"git.ronaksoft.com/river/msg/go/msg"
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

	r.Log().Debug("AccountModule applies UpdateAccountPrivacy",
		zap.Int64("UpdateID", x.UpdateID),
	)

	err = repo.Account.SetPrivacy(msg.PrivacyKey_PrivacyKeyChatInvite, x.ChatInvite)
	if err != nil {
		r.Log().Error("AccountModule got error on set privacy (ChatInvite)", zap.Error(err))
	}

	err = repo.Account.SetPrivacy(msg.PrivacyKey_PrivacyKeyLastSeen, x.LastSeen)
	if err != nil {
		r.Log().Error("AccountModule got error on set privacy (LastSeen)", zap.Error(err))
	}

	err = repo.Account.SetPrivacy(msg.PrivacyKey_PrivacyKeyPhoneNumber, x.PhoneNumber)
	if err != nil {
		r.Log().Error("AccountModule got error on set privacy (PhoneNumber)", zap.Error(err))
	}

	err = repo.Account.SetPrivacy(msg.PrivacyKey_PrivacyKeyProfilePhoto, x.ProfilePhoto)
	if err != nil {
		r.Log().Error("AccountModule got error on set privacy (ProfilePhoto)", zap.Error(err))
	}

	err = repo.Account.SetPrivacy(msg.PrivacyKey_PrivacyKeyForwardedMessage, x.ForwardedMessage)
	if err != nil {
		r.Log().Error("AccountModule got error on set privacy (ForwardedMessage)", zap.Error(err))
	}

	err = repo.Account.SetPrivacy(msg.PrivacyKey_PrivacyKeyCall, x.Call)
	r.Log().Error("AccountModule got error on set privacy (Call)", zap.Error(err))
	res := []*msg.UpdateEnvelope{u}
	return res, nil
}
