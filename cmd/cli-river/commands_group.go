package main

import (
    "github.com/abiosoft/ishell/v2"
    "github.com/ronaksoft/river-msg/go/msg"
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
        if reqID, err := _SDK.ExecuteCommand(msg.C_GroupsCreate, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }
    },
}

var AddUser = &ishell.Cmd{
    Name: "AddUser",
    Func: func(c *ishell.Context) {
        req := msg.GroupsAddUser{}
        req.User = &msg.InputUser{}
        req.GroupID = fnGetGroupID(c)
        req.User.UserID = fnGetPeerID(c)
        req.User.AccessHash = fnGetAccessHash(c)
        req.ForwardLimit = fnGetForwardLimit(c)
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        if reqID, err := _SDK.ExecuteCommand(msg.C_GroupsAddUser, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }
    },
}

var DeleteUser = &ishell.Cmd{
    Name: "DeleteUser",
    Func: func(c *ishell.Context) {
        req := msg.GroupsDeleteUser{}
        req.User = &msg.InputUser{}
        req.GroupID = fnGetGroupID(c)
        req.User.UserID = fnGetPeerID(c)
        req.User.AccessHash = fnGetAccessHash(c)
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        if reqID, err := _SDK.ExecuteCommand(msg.C_GroupsDeleteUser, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }
    },
}

var EditTitle = &ishell.Cmd{
    Name: "EditTitle",
    Func: func(c *ishell.Context) {
        req := msg.GroupsEditTitle{}
        req.GroupID = fnGetGroupID(c)
        req.Title = fnGetTitle(c)
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        if reqID, err := _SDK.ExecuteCommand(msg.C_GroupsEditTitle, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }
    },
}

var GetFull = &ishell.Cmd{
    Name: "GetFull",
    Func: func(c *ishell.Context) {
        req := msg.GroupsGetFull{}
        req.GroupID = fnGetGroupID(c)
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        _, err := _SDK.ExecuteCommand(msg.C_GroupsGetFull, reqBytes, reqDelegate)
        if err != nil {
            c.Println("Command Failed:", err)
        }
    },
}

var UpdateAdmin = &ishell.Cmd{
    Name: "UpdateAdmin",
    Func: func(c *ishell.Context) {
        req := msg.GroupsUpdateAdmin{}
        req.User = new(msg.InputUser)
        req.GroupID = fnGetGroupID(c)
        req.User.UserID = fnGetPeerID(c)
        req.User.AccessHash = fnGetAccessHash(c)
        req.Admin = fnGetAdmin(c)

        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        if reqID, err := _SDK.ExecuteCommand(msg.C_GroupsUpdateAdmin, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }
    },
}

var ToggleAdmins = &ishell.Cmd{
    Name: "ToggleAdmins",
    Func: func(c *ishell.Context) {
        req := msg.GroupsToggleAdmins{}
        req.GroupID = fnGetGroupID(c)
        req.AdminEnabled = fnGetAdminEnabled(c)
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        if reqID, err := _SDK.ExecuteCommand(msg.C_GroupsToggleAdmins, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }
    },
}

var ToggleAdminOnly = &ishell.Cmd{
    Name: "ToggleAdminOnly",
    Func: func(c *ishell.Context) {
        req := msg.GroupsToggleAdminOnly{}
        req.GroupID = fnGetGroupID(c)
        req.AdminOnly = fnGetBool(c, "AdminOnly")
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        if reqID, err := _SDK.ExecuteCommand(msg.C_GroupsToggleAdminOnly, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }
    },
}

var GroupUploadPhoto = &ishell.Cmd{
    Name: "UploadPhoto",
    Func: func(c *ishell.Context) {
        groupID := fnGetGroupID(c)
        filePath := fnGetFilePath(c)
        _SDK.GroupUploadPhoto(groupID, filePath)

    },
}

var GroupRemovePhoto = &ishell.Cmd{
    Name: "RemovePhoto",
    Func: func(c *ishell.Context) {
        req := new(msg.GroupsRemovePhoto)
        req.GroupID = fnGetGroupID(c)
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)

        if reqID, err := _SDK.ExecuteCommand(msg.C_GroupsRemovePhoto, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }
    },
}

func init() {
    Group.AddCmd(Create)
    Group.AddCmd(AddUser)
    Group.AddCmd(DeleteUser)
    Group.AddCmd(EditTitle)
    Group.AddCmd(GetFull)
    Group.AddCmd(UpdateAdmin)
    Group.AddCmd(ToggleAdmins)
    Group.AddCmd(ToggleAdminOnly)
    Group.AddCmd(GroupUploadPhoto)
    Group.AddCmd(GroupRemovePhoto)
}
