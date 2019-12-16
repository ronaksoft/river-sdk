package labelHole

import (
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
)

/*
   Creation Time: 2019 - Dec - 16
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

type Bar struct {
	MinID int64
	MaxID int64
}

func Fill(minID, maxID int64) error {
	b := GetFilled()
	if maxID > b.MaxID {
		_ = repo.System.SaveInt(domain.SkLabelMinID, uint64(minID))
	}

	if b.MinID == 0 || minID < b.MinID {
		_ = repo.System.SaveInt(domain.SkLabelMaxID, uint64(maxID))
	}

	return nil
}

func GetFilled() Bar {
	minID, _ := repo.System.LoadInt64(domain.SkLabelMinID)
	maxID, _ := repo.System.LoadInt64(domain.SkLabelMaxID)
	return Bar{
		MinID: minID,
		MaxID: maxID,
	}
}

func GetLowerFilled(maxID int64) (bool, Bar) {
	b := GetFilled()
	if maxID > b.MaxID || maxID < b.MinID {
		return false, Bar{}
	}
	b.MaxID = maxID
	return true, b
}

func GetUpperFilled(minID int64) (bool, Bar) {
	b := GetFilled()
	if minID < b.MinID || minID > b.MaxID {
		return false, Bar{}
	}
	b.MinID = minID
	return true, b
}
