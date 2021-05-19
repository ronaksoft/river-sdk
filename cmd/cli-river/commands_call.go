package main

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"gopkg.in/abiosoft/ishell.v2"
)

/*
   Creation Time: 2021 - May - 19
   Created by:  (Hamidrezakk)
   Maintainers:
      1.  Hamidrezakk
   Auditor: Hamidrezakk
   Copyright Ronak Software Group 2021
*/

var Call = &ishell.Cmd{
	Name: "Call",
}

var CallStart = &ishell.Cmd{
	Name: "Start",
	Func: func(c *ishell.Context) {
		req := msg.ClientCallStart{}
		req.CallID = 0
		req.Peer = &msg.InputPeer{}
		req.Peer.Type = fnGetPeerType(c)
		req.Peer.ID = fnGetPeerID(c)
		req.Peer.AccessHash = fnGetAccessHash(c)
		req.Video = false
		inputUser := &msg.InputUser{}
		inputUser.UserID = req.Peer.ID
		inputUser.AccessHash = req.Peer.AccessHash
		req.InputUsers = append(req.InputUsers, inputUser)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		// id: 1201478792185258
		// accesshash: 4502681147619876
		if reqID, err := _SDK.ExecuteCommand(msg.C_ClientCallStart, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

func init() {
	Call.AddCmd(CallStart)
}
