package label

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/module"
)

/*
   Creation Time: 2021 - May - 10
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

type label struct {
	module.Base
}

func New() *label {
	r := &label{}
	r.RegisterHandlers(
		map[int64]domain.LocalMessageHandler{
			msg.C_LabelsAddToMessage:      r.labelAddToMessage,
			msg.C_LabelsDelete:            r.labelsDelete,
			msg.C_LabelsGet:               r.labelsGet,
			msg.C_LabelsListItems:         r.labelsListItems,
			msg.C_LabelsRemoveFromMessage: r.labelRemoveFromMessage,
		},
	)
	return r
}
