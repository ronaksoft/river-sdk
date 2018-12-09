package main

import (
	"bytes"
	"fmt"

	"git.ronaksoftware.com/ronak/riversdk/msg"
	"github.com/olekukonko/tablewriter"
	"go.uber.org/zap"
)

func MessagePrinter(envelope *msg.MessageEnvelope) {
	constructorName, _ := msg.ConstructorNames[envelope.Constructor]
	_Shell.Println(_GREEN("ConstructorName: %s (0x%X)", constructorName, envelope.Constructor))
	switch envelope.Constructor {
	case msg.C_AuthAuthorization:
		x := new(msg.AuthAuthorization)
		x.Unmarshal(envelope.Message)
		_Shell.Println(_BLUE("%s %s (%d)", x.User.FirstName, x.User.LastName, x.User.ID))
	case msg.C_AuthCheckedPhone:
		x := new(msg.AuthCheckedPhone)
		x.Unmarshal(envelope.Message)
		_Shell.Println(_BLUE("Registered: %t", x.Registered))
	case msg.C_AuthRecalled:
		x := new(msg.AuthRecalled)
		x.Unmarshal(envelope.Message)
		_Shell.Println(_BLUE("ClientID: %d", x.ClientID))
	case msg.C_AuthSentCode:
		x := new(msg.AuthSentCode)
		x.Unmarshal(envelope.Message)
		_Shell.Println(_BLUE("Phone, PhoneCodeHash: %s, %s", x.Phone, x.PhoneCodeHash))
	case msg.C_ContactsImported:
		x := new(msg.ContactsImported)
		x.Unmarshal(envelope.Message)
		buf := new(bytes.Buffer)
		table := tablewriter.NewWriter(buf)
		table.SetHeader([]string{
			"FirstName", "LastName", "Username", "User ID", "AccessHash",
		})

		for _, u := range x.Users {
			table.Append([]string{
				u.FirstName,
				u.LastName,
				u.Username,
				fmt.Sprintf("%d", u.ID),
				fmt.Sprintf("%d", u.AccessHash),
			})
		}
		table.Render()
		_Shell.Println(buf.String())
	case msg.C_ContactsMany:
		x := new(msg.ContactsMany)
		x.Unmarshal(envelope.Message)

		bufUsers := new(bytes.Buffer)
		tableUsers := tablewriter.NewWriter(bufUsers)
		tableUsers.SetCaption(true, "Users")
		tableUsers.SetHeader([]string{
			"FirstName", "LastName", "Username", "User ID", "AccessHash", "ClientID",
		})

		for _, u := range x.Users {
			tableUsers.Append([]string{
				u.FirstName,
				u.LastName,
				u.Username,
				fmt.Sprintf("%d", u.ID),
				fmt.Sprintf("%d", u.AccessHash),
				fmt.Sprintf("%d", u.ClientID),
			})
		}
		tableUsers.Render()
		bufContacts := new(bytes.Buffer)
		tableContacts := tablewriter.NewWriter(bufContacts)
		tableContacts.SetCaption(true, "Contacts")
		tableContacts.SetHeader([]string{
			"Client ID", "FirstName", "LastName", "Phone", "ClientID",
		})

		for _, u := range x.Contacts {
			tableContacts.Append([]string{
				fmt.Sprintf("%d", u.ClientID),
				u.FirstName,
				u.LastName,
				u.Phone,
				fmt.Sprintf("%d", u.ClientID),
			})
		}
		tableContacts.Render()
		_Shell.Println(bufUsers.String())
		_Shell.Println(bufContacts.String())
	case msg.C_MessagesDialogs:
		x := new(msg.MessagesDialogs)
		x.Unmarshal(envelope.Message)
		_Shell.Println(_BLUE("Total: %d", x.Count))
		bufDialogs := new(bytes.Buffer)
		tableDialogs := tablewriter.NewWriter(bufDialogs)
		tableDialogs.SetHeader([]string{
			"PeerID", "PeerType", "Top Message ID", "Unread", "AccessHash", "Falgs",
		})

		for _, d := range x.Dialogs {
			tableDialogs.Append([]string{
				fmt.Sprintf("%d", d.PeerID),
				fmt.Sprintf("%d", d.PeerType),
				fmt.Sprintf("%d", d.TopMessageID),
				fmt.Sprintf("%d", d.UnreadCount),
				fmt.Sprintf("%d", d.AccessHash),
				fmt.Sprintf("%d", d.Flags),
			})
		}
		tableDialogs.Render()
		bufUsers := new(bytes.Buffer)
		tableUsers := tablewriter.NewWriter(bufUsers)
		tableUsers.SetHeader([]string{
			"UserID", "FirstName", "LastName",
		})
		for _, x := range x.Users {
			tableUsers.Append([]string{
				fmt.Sprintf("%d", x.ID),
				fmt.Sprintf("%s", x.FirstName),
				fmt.Sprintf("%s", x.LastName),
			})
		}
		tableUsers.Render()
		// group
		bufGroup := new(bytes.Buffer)
		tableGroup := tablewriter.NewWriter(bufGroup)
		tableGroup.SetHeader([]string{
			"GroupID", "Title",
		})
		for _, x := range x.Groups {
			tableGroup.Append([]string{
				fmt.Sprintf("%d", x.ID),
				fmt.Sprintf("%s", x.Title),
			})
		}
		tableGroup.Render()

		_Shell.Println(bufDialogs.String())
		_Shell.Println(bufUsers.String())
		_Shell.Println(bufGroup.String())
	case msg.C_Dialog:
		x := new(msg.Dialog)
		x.Unmarshal(envelope.Message)
		buf := new(bytes.Buffer)
		table := tablewriter.NewWriter(buf)
		table.SetHeader([]string{
			"PeerID", "PeerType", "Top Message ID", "Unread", "AccessHash",
		})
		table.Append([]string{
			fmt.Sprintf("%d", x.PeerID),
			fmt.Sprintf("%d", x.PeerType),
			fmt.Sprintf("%d", x.TopMessageID),
			fmt.Sprintf("%d", x.UnreadCount),
			fmt.Sprintf("%d", x.AccessHash),
		})
		table.Render()
		_Shell.Println(buf.String())

	case msg.C_MessagesSent:
		x := new(msg.MessagesSent)
		x.Unmarshal(envelope.Message)
		_Shell.Println(_BLUE("MessageID, RandomID: %d, %d", x.MessageID, x.RandomID))
	case msg.C_Bool:
		x := new(msg.Bool)
		x.Unmarshal(envelope.Message)
		_Shell.Println(_BLUE("Result: %t", x.Result))
	case msg.C_Error:
		x := new(msg.Error)
		x.Unmarshal(envelope.Message)
		_Shell.Println(_BLUE("%s:%s", x.Code, x.Items))
	case msg.C_MessagesMany:

		x := new(msg.MessagesMany)
		x.Unmarshal(envelope.Message)
		_Shell.Println(_BLUE("Total Message Count: %d", len(x.Messages)))
		bufMessages := new(bytes.Buffer)
		tableMessages := tablewriter.NewWriter(bufMessages)
		tableMessages.SetHeader([]string{
			"PeerID", "PeerType", "CreatedOn", "Flags", "Body",
		})

		for _, d := range x.Messages {
			tableMessages.Append([]string{
				fmt.Sprintf("%d", d.PeerID),
				fmt.Sprintf("%d", d.PeerType),
				fmt.Sprintf("%d", d.CreatedOn),
				fmt.Sprintf("%d", d.Flags),
				fmt.Sprintf("%v", string(d.Body)),
			})
		}
		tableMessages.Render()
		bufUsers := new(bytes.Buffer)
		tableUsers := tablewriter.NewWriter(bufUsers)
		tableUsers.SetHeader([]string{
			"UserID", "FirstName", "LastName",
		})
		for _, x := range x.Users {
			tableUsers.Append([]string{
				fmt.Sprintf("%d", x.ID),
				fmt.Sprintf("%s", x.FirstName),
				fmt.Sprintf("%s", x.LastName),
			})
		}
		tableUsers.Render()
		_Shell.Println(bufMessages.String())
		_Shell.Println(bufUsers.String())
	case msg.C_UsersMany:

		x := new(msg.UsersMany)
		x.Unmarshal(envelope.Message)
		bufUsers := new(bytes.Buffer)
		tableUsers := tablewriter.NewWriter(bufUsers)
		tableUsers.SetHeader([]string{
			"UserID", "FirstName", "LastName",
		})
		for _, x := range x.Users {
			tableUsers.Append([]string{
				fmt.Sprintf("%d", x.ID),
				fmt.Sprintf("%s", x.FirstName),
				fmt.Sprintf("%s", x.LastName),
			})
		}
		tableUsers.Render()
		_Shell.Println(bufUsers.String())
	case msg.C_UpdateDifference:
		x := new(msg.UpdateDifference)
		x.Unmarshal(envelope.Message)

		for _, v := range x.Updates {
			if v.Constructor == msg.C_UpdateNewMessage {
				msg := new(msg.MessageEnvelope)
				msg.Constructor = v.Constructor
				msg.Message = v.Update

				MessagePrinter(msg)
			}

		}
	case msg.C_UpdateNewMessage:
		x := new(msg.UpdateNewMessage)
		x.Unmarshal(envelope.Message)

		bufMsg := new(bytes.Buffer)
		tableMsg := tablewriter.NewWriter(bufMsg)
		tableMsg.SetHeader([]string{
			"UpdateID", "AccessHash", "Sender", "Message.ID", "Message.Body",
		})

		tableMsg.Append([]string{
			fmt.Sprintf("%d", x.UpdateID),
			fmt.Sprintf("%d", x.AccessHash),
			fmt.Sprintf("%s %s", x.Sender.FirstName, x.Sender.LastName),
			fmt.Sprintf("%d", x.Message.ID),
			fmt.Sprintf("%s", x.Message.Body),
		})

		tableMsg.Render()
		_Shell.Println(bufMsg.String())

	case msg.C_GroupFull:
		x := new(msg.GroupFull)
		err := x.Unmarshal(envelope.Message)
		if err != nil {
			_Shell.Println(_RED(err.Error()))
			return
		}
		if x.Group != nil {
			_Shell.Println(fmt.Sprintf("GroupID : %d \t Title : %s", x.Group.ID, x.Group.Title))
		} else {
			_Shell.Println(_RED("x.Group is null"))
		}
		if x.NotifySettings != nil {
			_Shell.Println(fmt.Sprintf("NotifySettings Sound: %s \t Mute : %d \t Flag : %d", x.NotifySettings.Sound, x.NotifySettings.MuteUntil, x.NotifySettings.Flags))
		} else {
			_Shell.Println(_RED("x.NotifySettings is null"))
		}
		if x.Participants != nil {
			_Shell.Println(fmt.Sprintf("Participants Count : %d ", len(x.Participants)))
		} else {
			_Shell.Println(_RED("x.Participants is null"))
		}

		bufUsers := new(bytes.Buffer)
		tableUsers := tablewriter.NewWriter(bufUsers)
		tableUsers.SetHeader([]string{
			"UserID", "FirstName", "LastName",
		})
		for _, x := range x.Users {
			tableUsers.Append([]string{
				fmt.Sprintf("%d", x.ID),
				fmt.Sprintf("%s", x.FirstName),
				fmt.Sprintf("%s", x.LastName),
			})
		}
		tableUsers.Render()
		_Shell.Println(bufUsers.String())
	default:
		constructorName, _ := msg.ConstructorNames[envelope.Constructor]
		_Log.Debug("DEFAULT",
			zap.String("ConstructorName", constructorName),
			zap.Int64("Constructor", envelope.Constructor),
		)
	}
}
