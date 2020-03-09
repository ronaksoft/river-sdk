package repo

import (
	"fmt"
	msg "git.ronaksoftware.com/river/msg/chat"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/dgraph-io/badger"
)

const (
	prefixAccount = "ACCOUNT"
)

type repoAccount struct {
	*repository
}

func (r *repoAccount) SetPrivacy(key msg.PrivacyKey, rules []*msg.PrivacyRule) error {
	accountPrivacyRules := new(msg.AccountPrivacyRules)
	accountPrivacyRules.Rules = rules

	bytes, _ := accountPrivacyRules.Marshal()
	err := badgerUpdate(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(
			ronak.StrToByte(fmt.Sprintf("%s.%s", prefixAccount, key)),
			bytes,
		))
	})
	return err
}

func (r *repoAccount) GetPrivacy(key msg.PrivacyKey) (*msg.AccountPrivacyRules, error) {
	var rulesBytes []byte
	err := badgerView(func(txn *badger.Txn) error {
		item, err := txn.Get(ronak.StrToByte(fmt.Sprintf("%s.%s", prefixAccount, key)))
		if err != nil {
			return err
		}
		rulesBytes, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	accountPrivacyRules := new(msg.AccountPrivacyRules)
	_ = accountPrivacyRules.Unmarshal(rulesBytes)
	return accountPrivacyRules, nil
}
