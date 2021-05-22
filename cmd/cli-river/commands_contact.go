package main

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	riversdk "git.ronaksoft.com/river/sdk/sdk/prime"
	"gopkg.in/abiosoft/ishell.v2"
)

var Contact = &ishell.Cmd{
	Name: "Contact",
}

var ContactImport = &ishell.Cmd{
	Name: "Import",
	Func: func(c *ishell.Context) {
		req := msg.ContactsImport{}
		req.Replace = true
		contact := msg.PhoneContact{}
		contact.FirstName = fnGetFirstName(c)
		contact.LastName = fnGetLastName(c)
		contact.Phone = fnGetPhone(c)
		contact.ClientID = domain.RandomInt63()
		req.Contacts = append(req.Contacts, &contact)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_ContactsImport, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var ContactGet = &ishell.Cmd{
	Name: "Get",
	Func: func(c *ishell.Context) {
		req := &msg.ContactsGet{}
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_ContactsGet, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var ContactGetTeam = &ishell.Cmd{
	Name: "GetTeam",
	Func: func(c *ishell.Context) {
		req := msg.ContactsGet{}
		teamID := fnGetTeamID(c)
		accessHash := fnGetAccessHash(c)
		_SDK.SetTeam(teamID, int64(accessHash), false)
		reqBytes, _ := req.Marshal()

		reqDelegate := NewCustomDelegate()
		reqDelegate.FlagsFunc = func() riversdk.RequestDelegateFlag {
			return domain.RequestServerForced
		}
		if reqID, err := _SDK.ExecuteCommand(msg.C_ContactsGet, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var ContactAdd = &ishell.Cmd{
	Name: "Add",
	Func: func(c *ishell.Context) {
		req := msg.ContactsAdd{}
		req.FirstName = fnGetFirstName(c)
		req.LastName = fnGetLastName(c)
		req.Phone = fnGetPhone(c)
		req.User = fnGetUser(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_ContactsAdd, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var ContactDelete = &ishell.Cmd{
	Name: "Delete",
	Func: func(c *ishell.Context) {
		req := msg.ContactsDelete{}
		req.UserIDs = append(req.UserIDs, fnGetUserID(c))
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_ContactsDelete, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var ContactSearch = &ishell.Cmd{
	Name: "Search",
	Func: func(c *ishell.Context) {
		req := msg.ContactsSearch{}
		req.Q = fnGetUsername(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_ContactsSearch, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var ContactGetTopPeers = &ishell.Cmd{
	Name: "GetTopPeers",
	Func: func(c *ishell.Context) {
		req := msg.ContactsGetTopPeers{
			Limit: 10,
		}
		req.Category = fnGetTopPeerCat(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_ContactsGetTopPeers, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var ContactDeleteAll = &ishell.Cmd{
	Name: "DeleteAll",
	Func: func(c *ishell.Context) {
		req := msg.ContactsDeleteAll{}
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_ContactsDeleteAll, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

func init() {
	Contact.AddCmd(ContactImport)
	Contact.AddCmd(ContactGet)
	Contact.AddCmd(ContactGetTeam)
	Contact.AddCmd(ContactAdd)
	Contact.AddCmd(ContactDelete)
	Contact.AddCmd(ContactGetTopPeers)
	Contact.AddCmd(ContactDeleteAll)
	Contact.AddCmd(ContactSearch)
}
