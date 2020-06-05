package main

import (
	"git.ronaksoftware.com/river/msg/msg"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
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
		req := msg.ContactsGet{}
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_ContactsGet, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var ContactTestImport = &ishell.Cmd{
	Name: "TestImport",
	Func: func(c *ishell.Context) {
		req := msg.ContactsImport{}
		req.Replace = true
		_, phoneContacts := _SDK.GetContacts()
		req.Contacts = phoneContacts
		// for i := 0 ; i < 1000; i++ {
		// 	contact := msg.PhoneContact{}
		// 	contact.FirstName = domain.RandomID(10)
		// 	contact.LastName = domain.RandomID(10)
		// 	contact.Phone = fmt.Sprintf("237400%04d", i)
		// 	contact.ClientID = domain.RandomInt63()
		// 	req.Contacts = append(req.Contacts, &contact)
		// }
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_ContactsImport, reqBytes, reqDelegate); err != nil {
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
		req.FirstName =  fnGetFirstName(c)
		req.LastName = fnGetLastName(c)
		req.Phone= fnGetPhone(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_ContactsAdd, reqBytes, reqDelegate); err != nil {
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
func init() {
	Contact.AddCmd(ContactImport)
	Contact.AddCmd(ContactGet)
	Contact.AddCmd(ContactTestImport)
	Contact.AddCmd(ContactAdd)
	Contact.AddCmd(ContactGetTopPeers)
}
