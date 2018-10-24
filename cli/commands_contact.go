package main

import (
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/toolbox"
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
		c.Print("First Name:")
		contact.FirstName = c.ReadLine()
		c.Print("Last Name:")
		contact.LastName = c.ReadLine()
		c.Print("Phone: ")
		contact.Phone = c.ReadLine()
		contact.ClientID = ronak.RandomInt64(0)
		req.Contacts = append(req.Contacts, &contact)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(int64(msg.C_ContactsImport), reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var ContactGet = &ishell.Cmd{
	Name: "Get",
	Func: func(c *ishell.Context) {
		req := msg.ContactsGet{}
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(int64(msg.C_ContactsGet), reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

func init() {
	Contact.AddCmd(ContactImport)
	Contact.AddCmd(ContactGet)
}
