package main

import (
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"go.uber.org/zap"
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
		if reqID, err := _SDK.ExecuteCommand(msg.C_GroupsCreate, reqBytes, reqDelegate, false, false); err != nil {
			logs.Error("ExecuteCommand failed", zap.Error(err))
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
		if reqID, err := _SDK.ExecuteCommand(msg.C_GroupsAddUser, reqBytes, reqDelegate, false, false); err != nil {
			logs.Error("ExecuteCommand failed", zap.Error(err))
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
		if reqID, err := _SDK.ExecuteCommand(msg.C_GroupsDeleteUser, reqBytes, reqDelegate, false, false); err != nil {
			logs.Error("ExecuteCommand failed", zap.Error(err))
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
		if reqID, err := _SDK.ExecuteCommand(msg.C_GroupsEditTitle, reqBytes, reqDelegate, false, false); err != nil {
			logs.Error("ExecuteCommand failed", zap.Error(err))
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
		if reqID, err := _SDK.ExecuteCommand(msg.C_GroupsGetFull, reqBytes, reqDelegate, false, false); err != nil {
			logs.Error("ExecuteCommand failed", zap.Error(err))
		} else {
			reqDelegate.RequestID = reqID
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
		if reqID, err := _SDK.ExecuteCommand(msg.C_GroupsUpdateAdmin, reqBytes, reqDelegate, false, false); err != nil {
			logs.Error("ExecuteCommand failed", zap.Error(err))
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
		if reqID, err := _SDK.ExecuteCommand(msg.C_GroupsToggleAdmins, reqBytes, reqDelegate, false, false); err != nil {
			logs.Error("ExecuteCommand failed", zap.Error(err))
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

var GroupDownloadPhotoBig = &ishell.Cmd{
	Name: "DownloadPhotoBig",
	Func: func(c *ishell.Context) {
		groupID := fnGetGroupID(c)
		strFilePath := _SDK.GroupGetPhoto_Big(groupID)
		logs.Info("File Download Complete", zap.String("path", strFilePath))

	},
}
var GroupDownloadPhotoSmall = &ishell.Cmd{
	Name: "DownloadPhotoSmall",
	Func: func(c *ishell.Context) {
		groupID := fnGetGroupID(c)
		strFilePath := _SDK.GroupGetPhoto_Small(groupID)
		logs.Info("File Download Complete", zap.String("path", strFilePath))
	},
}

var GroupRemovePhoto = &ishell.Cmd{
	Name: "RemovePhoto",
	Func: func(c *ishell.Context) {
		req := new(msg.GroupsRemovePhoto)
		req.GroupID = fnGetGroupID(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)

		if reqID, err := _SDK.ExecuteCommand(msg.C_GroupsRemovePhoto, reqBytes, reqDelegate, false, false); err != nil {
			logs.Error("ExecuteCommand failed", zap.Error(err))
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
	Group.AddCmd(GroupUploadPhoto)
	Group.AddCmd(GroupDownloadPhotoBig)
	Group.AddCmd(GroupDownloadPhotoSmall)
	Group.AddCmd(GroupRemovePhoto)
}
