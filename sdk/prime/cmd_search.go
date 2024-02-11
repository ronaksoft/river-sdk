package riversdk

import (
    "sync"

    "github.com/ronaksoft/river-sdk/internal/repo"
)

/*
   Creation Time: 2019 - Jul - 24
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func (r *River) SearchReIndex(teamID int64) {
    waitGroup := sync.WaitGroup{}
    waitGroup.Add(2)
    go func() {
        _ = repo.Users.ReIndex(teamID)
        _ = repo.Groups.ReIndex()
        waitGroup.Done()
    }()
    go func() {
        repo.Messages.ReIndex()
        waitGroup.Done()
    }()
    waitGroup.Wait()
}
