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
		req.Peer = fnGetPeer(c)
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

var CallReject = &ishell.Cmd{
	Name: "Reject",
	Func: func(c *ishell.Context) {
		req := msg.ClientCallReject{}
		req.CallID = fnGetCallID(c)
		req.Duration = 0
		req.Reason = msg.DiscardReason_DiscardReasonHangup
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_ClientCallReject, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var CallAccept = &ishell.Cmd{
	Name: "Accept",
	Func: func(c *ishell.Context) {
		req := msg.ClientCallAccept{}
		req.CallID = fnGetCallID(c)
		req.Video = false
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_ClientCallAccept, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var CallMediaSettings = &ishell.Cmd{
	Name: "MediaSettings",
	Func: func(c *ishell.Context) {
		req := msg.ClientCallSendMediaSettings{}
		req.MediaSettings = &msg.CallMediaSettings{
			Audio:       true,
			ScreenShare: false,
			Video:       false,
		}
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_ClientCallSendMediaSettings, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var CallIceCandidate = &ishell.Cmd{
	Name: "IceCandidate",
	Func: func(c *ishell.Context) {
		req := msg.ClientCallSendIceCandidate{}
		req.ConnId = fnGetConnId(c)
		req.Candidate  = &msg.CallRTCIceCandidate{
			Candidate:        "",
			SdpMLineIndex:    0,
			SdpMid:           "",
			UsernameFragment: "",
		}
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_ClientCallSendIceCandidate, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

func init() {
	Call.AddCmd(CallStart)
	Call.AddCmd(CallReject)
	Call.AddCmd(CallAccept)
	Call.AddCmd(CallMediaSettings)
	Call.AddCmd(CallIceCandidate)
}
