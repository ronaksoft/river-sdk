package label

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/module"
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
		map[int64]domain.LocalHandler{
			msg.C_LabelsAddToMessage:      r.labelAddToMessage,
			msg.C_LabelsDelete:            r.labelsDelete,
			msg.C_LabelsGet:               r.labelsGet,
			msg.C_LabelsListItems:         r.labelsListItems,
			msg.C_LabelsRemoveFromMessage: r.labelRemoveFromMessage,
		},
	)
	r.RegisterUpdateAppliers(
		map[int64]domain.UpdateApplier{
			msg.C_UpdateLabelDeleted:      r.updateLabelDeleted,
			msg.C_UpdateLabelItemsAdded:   r.updateLabelItemsAdded,
			msg.C_UpdateLabelItemsRemoved: r.updateLabelItemsRemoved,
			msg.C_UpdateLabelSet:          r.updateLabelSet,
		},
	)
	r.RegisterMessageAppliers(
		map[int64]domain.MessageApplier{
			msg.C_LabelItems: r.labelItems,
			msg.C_LabelsMany: r.labelsMany,
		},
	)
	return r
}

func (r *label) Name() string {
	return module.Label
}
