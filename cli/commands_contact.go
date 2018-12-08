package main

import (
	"git.ronaksoftware.com/ronak/riversdk/domain"
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
		contact.FirstName = fnGetFirstName(c)
		contact.LastName = fnGetLastName(c)
		contact.Phone = fnGetPhone(c)
		contact.ClientID = ronak.RandomInt64(0)
		req.Contacts = append(req.Contacts, &contact)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_ContactsImport, reqBytes, reqDelegate, false, false); err != nil {
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
		if reqID, err := _SDK.ExecuteCommand(msg.C_ContactsGet, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var SearchContacts = &ishell.Cmd{
	Name: "SearchContacts",
	Func: func(c *ishell.Context) {

		reqID := uint64(domain.SequentialUniqueID())
		searchPharase := fnGetSearchPhrase(c)
		reqDelegate := new(RequestDelegate)
		reqDelegate.RequestID = int64(reqID)
		_SDK.SearchContacts(reqID, searchPharase, reqDelegate)
	},
}

func init() {
	Contact.AddCmd(ContactImport)
	Contact.AddCmd(ContactGet)
	Contact.AddCmd(SearchContacts)
}
