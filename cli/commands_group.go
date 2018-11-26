package main

import (
	"git.ronaksoftware.com/ronak/riversdk/msg"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

var Group = &ishell.Cmd{
	Name: "Group",
}

var Create = &ishell.Cmd{
	Name: "Create",
	Func: func(c *ishell.Context) {
		req := msg.GroupsCreate{}
		req.Title = fnGetTitle(c)
		req.Users = fnGetInputUser(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_GroupsCreate, reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var AddGroupUser = &ishell.Cmd{
	Name: "AddGroupUser",
	Func: func(c *ishell.Context) {
		req := msg.GroupsAddUser{}
		req.User = &msg.InputUser{}
		req.GroupID = fnGetGroupID(c)
		req.User.UserID = fnGetPeerID(c)
		req.User.AccessHash = fnGetAccessHash(c)
		req.ForwardLimit = fnGetForwardLimit(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_GroupsAddUser, reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var DeleteGroupUser = &ishell.Cmd{
	Name: "DeleteGroupUser",
	Func: func(c *ishell.Context) {
		req := msg.GroupsDeleteUser{}
		req.User = &msg.InputUser{}
		req.GroupID = fnGetGroupID(c)
		req.User.UserID = fnGetPeerID(c)
		req.User.AccessHash = fnGetAccessHash(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_GroupsDeleteUser, reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var EditGroupTitle = &ishell.Cmd{
	Name: "EditGroupTitle",
	Func: func(c *ishell.Context) {
		req := msg.GroupsEditTitle{}
		req.GroupID = fnGetGroupID(c)
		req.Title = fnGetTitle(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_GroupsEditTitle, reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var GetFullGroup = &ishell.Cmd{
	Name: "GetFullGroup",
	Func: func(c *ishell.Context) {
		req := msg.GroupsGetFull{}
		req.GroupID = fnGetGroupID(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_GroupsGetFull, reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

func init() {
	Group.AddCmd(Create)
	Group.AddCmd(AddGroupUser)
	Group.AddCmd(DeleteGroupUser)
	Group.AddCmd(EditGroupTitle)
	Group.AddCmd(GetFullGroup)
}
