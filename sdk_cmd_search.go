package riversdk

import (
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
)

/*
   Creation Time: 2019 - Jul - 24
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/


func (r *River) SearchReIndex() {
	repo.Users.ReIndex()
	repo.Groups.ReIndex()
	repo.Messages.ReIndex()
}
