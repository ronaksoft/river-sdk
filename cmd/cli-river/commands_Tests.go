package main

import (
	"fmt"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"go.uber.org/zap"
	"gopkg.in/abiosoft/ishell.v2"
)

var Tests = &ishell.Cmd{
	Name: "Tests",
}

var BatchSend = &ishell.Cmd{
	Name: "BatchSend",
	Func: func(c *ishell.Context) {

		peerID := fnGetPeerID(c)
		accessHash := fnGetAccessHash(c)
		tries := fnGetTries(c)

		req := new(msg.MessageContainer)
		req.Length = int32(tries)
		req.Envelopes = make([]*msg.MessageEnvelope, tries)
		for i := 0; i < tries; i++ {
			m := new(msg.MessagesSend)
			m.Peer = &msg.InputPeer{
				ID:         peerID,
				AccessHash: accessHash,
				Type:       1,
			}
			m.RandomID = domain.SequentialUniqueID()
			m.Body = fmt.Sprintf("%d", i)

			e := new(msg.MessageEnvelope)
			e.Constructor = msg.C_MessagesSend
			e.Message, _ = m.Marshal()
			e.RequestID = uint64(domain.SequentialUniqueID())

			req.Envelopes[i] = e
		}

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(int64(msg.C_MessageContainer), reqBytes, reqDelegate, false, false); err != nil {
			_Log.Error("EnqueueCommand failed", zap.Error(err))
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

func init() {
	Tests.AddCmd(BatchSend)

}
