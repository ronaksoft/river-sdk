package salt

import (
	"encoding/json"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"github.com/dgraph-io/badger/v2"
	"github.com/ronaksoft/rony/tools"
	"go.uber.org/zap"
	"sort"
	"time"
)

/*
   Creation Time: 2019 - Jul - 13
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software GroupSearch 2018
*/

var (
	salts   []domain.Slt
	curSalt int64
	logger  *logs.Logger
)

func init() {
	logger = logs.With("SALT")
}

func Get() int64 {
	return curSalt
}

func Reset() {
	_ = repo.System.Delete(domain.SkSystemSalts)
	curSalt = 0
	UpdateSalt()
}

func UpdateSalt() bool {
	// 1st try to load from already stored salts
	saltString, err := repo.System.LoadString(domain.SkSystemSalts)
	if err != nil {
		switch err {
		case badger.ErrKeyNotFound:
			logger.Warn("UpdateSalt did not find salt key in the db")
		default:
			logger.Warn("UpdateSalt got error on load salt from db", zap.Error(err))

		}
		return false
	}

	var sysSalts []domain.Slt
	err = json.Unmarshal([]byte(saltString), &sysSalts)
	if err != nil {
		logger.Warn("UpdateSalt got error on unmarshal salt from db", zap.Error(err))
		return false
	}

	if len(sysSalts) > 0 {
		sort.Slice(sysSalts, func(i, j int) bool {
			return sysSalts[i].Timestamp < sysSalts[j].Timestamp
		})

		saltFound := false
		for idx, s := range sysSalts {
			validUntil := s.Timestamp + int64(time.Hour/time.Second) - domain.Now().Unix()
			if validUntil <= 0 {
				logger.Debug("did not match", zap.Any("salt timestamp", s.Timestamp))
				continue
			}
			curSalt = s.Value
			salts = sysSalts[idx:]
			b, _ := json.Marshal(sysSalts[idx:])
			err = repo.System.SaveString(domain.SkSystemSalts, string(b))
			if err != nil {
				logger.Warn("UpdateSalt got error on save salt to db",
					zap.Error(err),
					zap.String("Salts", tools.ByteToStr(b)),
				)
			}
			saltFound = true
			break
		}
		if !saltFound {
			logger.Warn("UpdateSalt could not find salt")
			return false
		}
		return true
	} else {
		logger.Warn("UpdateSalt could not find salt (length is zero)")
		return false
	}

}

func Set(s *msg.SystemSalts) {
	var saltArray []domain.Slt
	for idx, saltValue := range s.Salts {
		slt := domain.Slt{}
		slt.Timestamp = s.StartsFrom + (s.Duration/int64(time.Second))*int64(idx)
		slt.Value = saltValue
		saltArray = append(saltArray, slt)
	}
	b, _ := json.Marshal(saltArray)
	err := repo.System.SaveString(domain.SkSystemSalts, string(b))
	if err != nil {
		logger.Error("couldn't save SystemSalts in the db", zap.Error(err))
		return
	}
	UpdateSalt()
}

func Count() int {
	return len(salts)
}
