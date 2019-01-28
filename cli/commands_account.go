package main

import (
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"go.uber.org/zap"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

var Account = &ishell.Cmd{
	Name: "Account",
}

var RegisterDevice = &ishell.Cmd{
	Name: "RegisterDevice",
	Func: func(c *ishell.Context) {
		req := msg.AccountRegisterDevice{}
		req.TokenType = fnGetTokenType(c)
		req.Token = fnGetToken(c)
		req.DeviceModel = fnGetDeviceModel(c)
		req.SystemVersion = fnGetSysytemVersion(c)
		req.AppVersion = fnGetAppVersion(c)
		req.LangCode = fnGetLangCode(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)

		if reqID, err := _SDK.ExecuteCommand(msg.C_AccountRegisterDevice, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var UpdateUsername = &ishell.Cmd{
	Name: "UpdateUsername",
	Func: func(c *ishell.Context) {
		req := msg.AccountUpdateUsername{}
		req.Username = fnGetUsername(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)

		if reqID, err := _SDK.ExecuteCommand(msg.C_AccountUpdateUsername, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var CheckUsername = &ishell.Cmd{
	Name: "CheckUsername",
	Func: func(c *ishell.Context) {
		req := msg.AccountCheckUsername{}
		req.Username = fnGetUsername(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)

		if reqID, err := _SDK.ExecuteCommand(msg.C_AccountCheckUsername, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var UnregisterDevice = &ishell.Cmd{
	Name: "UnregisterDevice",
	Func: func(c *ishell.Context) {
		req := msg.AccountUnregisterDevice{}
		req.TokenType = 1
		req.Token = "token"

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)

		if reqID, err := _SDK.ExecuteCommand(msg.C_AccountUnregisterDevice, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var UpdateProfile = &ishell.Cmd{
	Name: "UpdateProfile",
	Func: func(c *ishell.Context) {
		req := msg.AccountUpdateProfile{}
		req.FirstName = fnGetFirstName(c)
		req.LastName = fnGetLastName(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)

		if reqID, err := _SDK.ExecuteCommand(msg.C_AccountUpdateProfile, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var SetNotifySettings = &ishell.Cmd{
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

		if reqID, err := _SDK.ExecuteCommand(msg.C_AccountSetNotifySettings, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var UploadPhoto = &ishell.Cmd{
	Name: "UploadPhoto",
	Func: func(c *ishell.Context) {
		filePath := fnGetFilePath(c)
		_SDK.AccountUploadPhoto(filePath)

	},
}

var DownloadPhotoBig = &ishell.Cmd{
	Name: "DownloadPhotoBig",
	Func: func(c *ishell.Context) {
		userID := fnGetPeerID(c)
		strFilePath := _SDK.AccountGetPhoto_Big(userID)
		log.LOG_Info("File Download Complete", zap.String("path", strFilePath))

	},
}
var DownloadPhotoSmall = &ishell.Cmd{
	Name: "DownloadPhotoSmall",
	Func: func(c *ishell.Context) {
		userID := fnGetPeerID(c)
		strFilePath := _SDK.AccountGetPhoto_Small(userID)
		log.LOG_Info("File Download Complete", zap.String("path", strFilePath))

	},
}

func init() {
	Account.AddCmd(RegisterDevice)
	Account.AddCmd(UpdateUsername)
	Account.AddCmd(CheckUsername)
	Account.AddCmd(UnregisterDevice)
	Account.AddCmd(UpdateProfile)
	Account.AddCmd(SetNotifySettings)
	Account.AddCmd(UploadPhoto)
	Account.AddCmd(DownloadPhotoBig)
	Account.AddCmd(DownloadPhotoSmall)
}
