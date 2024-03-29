package main

import (
    "github.com/abiosoft/ishell/v2"
    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/request"
    riversdk "github.com/ronaksoft/river-sdk/sdk/prime"
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
        ah := fnGetAccessHash(c)
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        if reqID, err := _SDK.ExecuteCommandWithTeam(req.TeamID, int64(ah), msg.C_TeamListMembers, reqBytes, reqDelegate); err != nil {
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
        reqDelegate.FlagsFunc = func() riversdk.RequestDelegateFlag {
            return request.ServerForced
        }

        if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesGetDialogs, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }
    },
}

var TeamEdit = &ishell.Cmd{
    Name: "Edit",
    Func: func(c *ishell.Context) {
        req := msg.TeamEdit{}
        req.TeamID = fnGetTeamID(c)
        req.Name = fnGetString(c, "Name")
        teamID := req.TeamID
        accesshHash := fnGetAccessHash(c)
        _SDK.SetTeam(teamID, int64(accesshHash), false)
        reqBytes, _ := req.Marshal()
        reqDelegate := NewCustomDelegate()
        reqDelegate.FlagsFunc = func() riversdk.RequestDelegateFlag {
            return request.ServerForced
        }

        if reqID, err := _SDK.ExecuteCommand(msg.C_TeamEdit, reqBytes, reqDelegate); err != nil {
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
    Team.AddCmd(TeamEdit)
}
