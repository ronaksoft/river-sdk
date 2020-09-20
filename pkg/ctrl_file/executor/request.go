package executor

import (
	"context"
)

/*
   Creation Time: 2020 - Sep - 20
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

type Action interface {
	ID() int32
	Do(ctx context.Context)
}

type Request interface {
	Prepare()
	NextAction() Action
	ActionDone(id int32)
	Serialize() []byte
	Deserialize([]byte)
}
