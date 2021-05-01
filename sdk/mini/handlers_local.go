package mini

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"github.com/ronaksoft/rony"
)

/*
   Creation Time: 2021 - May - 01
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func (r *River) messagesSendMedia(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesSendMedia{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	req.RandomID = domain.SequentialUniqueID()
	requestBytes, _ := req.Marshal()
	r.networkCtrl.HttpCommand(
		&rony.MessageEnvelope{
			Constructor: msg.C_MessagesSendMedia,
			RequestID:   uint64(req.RandomID),
			Message:     requestBytes,
			Header:      in.Header,
		}, timeoutCB, successCB,
	)
}
