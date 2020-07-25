package main

import (
	"git.ronaksoftware.com/river/msg/msg"
	"gopkg.in/abiosoft/ishell.v2"
)

var Team = &ishell.Cmd{
	Name: "Team",
}

var TeamAddMember = &ishell.Cmd{
	Name: "AddMember",
	Func: func(c *ishell.Context) {
		req := msg.TeamAddMember{}
		req.TeamID = fnGetTeamID(c)
		req.UserID = fnGetUserID(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_TeamAddMember, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var TeamRemoveMember = &ishell.Cmd{
	Name: "RemoveMember",
	Func: func(c *ishell.Context) {
		req := msg.TeamRemoveMember{}
		req.TeamID = fnGetTeamID(c)
		req.UserID = fnGetUserID(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_TeamRemoveMember, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}
var TeamListMembers = &ishell.Cmd{
	Name: "ListMembers",
	Func: func(c *ishell.Context) {
		req := msg.TeamListMembers{}
		req.TeamID = fnGetTeamID(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_TeamListMembers, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

func init() {
	Team.AddCmd(TeamAddMember)
	Team.AddCmd(TeamRemoveMember)
	Team.AddCmd(TeamListMembers)
}
