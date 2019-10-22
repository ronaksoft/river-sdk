package salt

import (
	"encoding/json"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	ronak "git.ronaksoftware.com/ronak/toolbox"
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
	salts      []domain.Slt
	curSalt    int64
	lastUpdate time.Time
)

func Get() int64 {
	return curSalt
}

func UpdateSalt() bool {
	// 1st try to load from already stored salts
	saltString, err := repo.System.LoadString(domain.SkSystemSalts)
	if err != nil {
		logs.Warn("UpdateSalt got error on load salt from db", zap.Error(err))
		return false
	}

	var sysSalts []domain.Slt
	err = json.Unmarshal([]byte(saltString), &sysSalts)
	if err != nil {
		logs.Warn("UpdateSalt got error on load salt from db", zap.Error(err))
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
				logs.Debug("did not match", zap.Any("salt timestamp", s.Timestamp))
				continue
			}
			curSalt = s.Value
			salts = sysSalts[idx:]
			b, _ := json.Marshal(sysSalts[idx:])
			err = repo.System.SaveString(domain.SkSystemSalts, string(b))
			if err != nil {
				logs.Warn("UpdateSalt got error on save salt to db", zap.Error(err),zap.String("Salts", ronak.ByteToStr(b)))
			}
			saltFound = true
			lastUpdate = time.Now()
			break
		}
		if !saltFound {
			logs.Warn("UpdateSalt could not find salt")
			return false
		}
		return true
	} else {
		logs.Warn("UpdateSalt could not find salt (length is zero)")
		return false
	}

}
