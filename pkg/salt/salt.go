package salt

import (
	"encoding/json"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
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
   Copyright Ronak Software Group 2018
*/

var (
	salts      []domain.Slt
	curSalt    int64
	timeDiff   int64
	lastUpdate time.Time
)

func SetTimeDifference(diff int64) {
	timeDiff = diff
}

func serverTime() time.Time {
	return time.Now().Add(time.Duration(timeDiff) * time.Second)
}

func Get() int64 {
	return curSalt
}

func UpdateSalt() bool {
	// 1st try to load from already stored salts
	saltString, err := repo.System.LoadString(domain.ColumnSystemSalts)
	if err != nil {
		return false
	}

	var sysSalts []domain.Slt
	err = json.Unmarshal([]byte(saltString), &sysSalts)
	if err != nil {
		return false
	}

	if len(sysSalts) > 0 {
		sort.Slice(sysSalts, func(i, j int) bool {
			return sysSalts[i].Timestamp < sysSalts[j].Timestamp
		})

		saltFound := false
		for idx, s := range sysSalts {
			validUntil := s.Timestamp + int64(time.Hour/time.Second) - serverTime().Unix()
			if validUntil <= 0 {
				logs.Debug("did not match", zap.Any("salt timestamp", s.Timestamp))
				continue
			}
			curSalt = s.Value
			salts = sysSalts[idx:]
			b, _ := json.Marshal(sysSalts[idx:])
			_ = repo.System.SaveString(domain.ColumnSystemSalts, string(b))
			saltFound = true
			lastUpdate = time.Now()
			break
		}
		if !saltFound {
			return false
		}
		return true
	} else {
		return false
	}

}
