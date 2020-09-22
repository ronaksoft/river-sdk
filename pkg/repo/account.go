package repo

import (
	"fmt"
	"git.ronaksoft.com/river/msg/msg"
	"git.ronaksoft.com/river/sdk/pkg/domain"
	"github.com/dgraph-io/badger/v2"
)

const (
	prefixAccount = "ACCOUNT"
)

type repoAccount struct {
	*repository
}

func (r *repoAccount) SetPrivacy(key msg.PrivacyKey, rules []*msg.PrivacyRule) error {
	accountPrivacyRules := &msg.AccountPrivacyRules{}
	accountPrivacyRules.Rules = rules

	bytes, _ := accountPrivacyRules.Marshal()
	err := badgerUpdate(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(
			domain.StrToByte(fmt.Sprintf("%s.%s", prefixAccount, key)),
			bytes,
		))
	})
	return err
}

func (r *repoAccount) GetPrivacy(key msg.PrivacyKey) (*msg.AccountPrivacyRules, error) {
	var rulesBytes []byte
	err := badgerView(func(txn *badger.Txn) error {
		item, err := txn.Get(domain.StrToByte(fmt.Sprintf("%s.%s", prefixAccount, key)))
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

	accountPrivacyRules := &msg.AccountPrivacyRules{}
	err = accountPrivacyRules.Unmarshal(rulesBytes)
	if err != nil {
		return nil, err
	}
	return accountPrivacyRules, nil
}
