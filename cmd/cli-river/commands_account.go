package main

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"gopkg.in/abiosoft/ishell.v2"
)

var Account = &ishell.Cmd{
	Name: "Account",
}

var AccountRegisterDevice = &ishell.Cmd{
	Name: "RegisterDevice",
	Func: func(c *ishell.Context) {
		req := msg.AccountRegisterDevice{}
		req.TokenType = fnGetTokenType(c)
		req.Token = fnGetToken(c)
		req.DeviceModel = fnGetDeviceModel(c)
		req.SystemVersion = fnGetSysytemVersion(c)
		req.AppVersion = fnGetAppVersion(c)
		req.LangCode = fnGetLangCode(c)
		req.ClientID = fnGetClientID(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)

		if reqID, err := _SDK.ExecuteCommand(msg.C_AccountRegisterDevice, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var AccountUpdateUsername = &ishell.Cmd{
	Name: "UpdateUsername",
	Func: func(c *ishell.Context) {
		req := msg.AccountUpdateUsername{}
		req.Username = fnGetUsername(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)

		if reqID, err := _SDK.ExecuteCommand(msg.C_AccountUpdateUsername, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var AccountCheckUsername = &ishell.Cmd{
	Name: "CheckUsername",
	Func: func(c *ishell.Context) {
		req := msg.AccountCheckUsername{}
		req.Username = fnGetUsername(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)

		if reqID, err := _SDK.ExecuteCommand(msg.C_AccountCheckUsername, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var AccountUnregisterDevice = &ishell.Cmd{
	Name: "UnregisterDevice",
	Func: func(c *ishell.Context) {
		req := msg.AccountUnregisterDevice{}
		req.TokenType = 1
		req.Token = "token"

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)

		if reqID, err := _SDK.ExecuteCommand(msg.C_AccountUnregisterDevice, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var AccountUpdateProfile = &ishell.Cmd{
	Name: "UpdateProfile",
	Func: func(c *ishell.Context) {
		req := msg.AccountUpdateProfile{}
		req.FirstName = fnGetFirstName(c)
		req.LastName = fnGetLastName(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_AccountUpdateProfile, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var AccountSetNotifySettings = &ishell.Cmd{
	Name: "SetNotifySettings",
	Func: func(c *ishell.Context) {
		req := msg.AccountSetNotifySettings{
			Peer:     new(msg.InputPeer),
			Settings: new(msg.PeerNotifySettings),
		}
		req.Peer.ID = fnGetPeerID(c)
		req.Peer.Type = 1
		req.Peer.AccessHash = fnGetAccessHash(c)
		req.Settings.Flags = 113
		req.Settings.MuteUntil = 0
		req.Settings.Sound = ""

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)

		if reqID, err := _SDK.ExecuteCommand(msg.C_AccountSetNotifySettings, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var AccountUploadPhoto = &ishell.Cmd{
	Name: "UploadPhoto",
	Func: func(c *ishell.Context) {
		filePath := fnGetFilePath(c)
		_SDK.AccountUploadPhoto(filePath)

	},
}

var AccountRemovePhoto = &ishell.Cmd{
	Name: "RemovePhoto",
	Func: func(c *ishell.Context) {
		req := msg.AccountRemovePhoto{}
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)

		if reqID, err := _SDK.ExecuteCommand(msg.C_AccountRemovePhoto, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var AccountGetTeams = &ishell.Cmd{
	Name: "GetTeams",
	Func: func(c *ishell.Context) {
		req := msg.AccountGetTeams{}
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)

		if reqID, err := _SDK.ExecuteCommand(msg.C_AccountGetTeams, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

func init() {
	Account.AddCmd(AccountRegisterDevice)
	Account.AddCmd(AccountUpdateUsername)
	Account.AddCmd(AccountCheckUsername)
	Account.AddCmd(AccountUnregisterDevice)
	Account.AddCmd(AccountUpdateProfile)
	Account.AddCmd(AccountSetNotifySettings)
	Account.AddCmd(AccountUploadPhoto)
	Account.AddCmd(AccountRemovePhoto)
	Account.AddCmd(AccountGetTeams)
}
