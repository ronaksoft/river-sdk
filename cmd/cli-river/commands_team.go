package main

import (
	"git.ronaksoft.com/river/msg/msg"
	"git.ronaksoft.com/ronak/riversdk"
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

var TeamPromote = &ishell.Cmd{
	Name: "Promote",
	Func: func(c *ishell.Context) {
		req := msg.TeamPromote{}
		req.TeamID = fnGetTeamID(c)
		req.UserID = fnGetUserID(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_TeamPromote, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var TeamDemote = &ishell.Cmd{
	Name: "Demote",
	Func: func(c *ishell.Context) {
		req := msg.TeamDemote{}
		req.TeamID = fnGetTeamID(c)
		req.UserID = fnGetUserID(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_TeamPromote, reqBytes, reqDelegate); err != nil {
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

var TeamGetDialogs = &ishell.Cmd{
	Name: "GetDialogs",
	Func: func(c *ishell.Context) {
		req := msg.MessagesGetDialogs{}
		req.Offset = fnGetOffset(c)
		req.Limit = 100
		teamID := fnGetTeamID(c)
		accesshHash := fnGetAccessHash(c)
		_SDK.SetTeam(teamID, int64(accesshHash), false)
		reqBytes, _ := req.Marshal()
		reqDelegate := NewCustomDelegate()
		reqDelegate.FlagsFunc = func() int32 {
			return riversdk.RequestServerForced
		}

		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesGetDialogs, reqBytes, reqDelegate); err != nil {
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
	Team.AddCmd(TeamPromote)
	Team.AddCmd(TeamDemote)
	Team.AddCmd(TeamGetDialogs)
}
