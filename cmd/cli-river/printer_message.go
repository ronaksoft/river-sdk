package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/msg/chat"
	"github.com/olekukonko/tablewriter"
	"go.uber.org/zap"
)

func MessagePrinter(envelope *msg.MessageEnvelope) {
	switch envelope.Constructor {
	case msg.C_AuthAuthorization:
		x := new(msg.AuthAuthorization)
		x.Unmarshal(envelope.Message)
		_Shell.Println(fmt.Sprintf("AuthAuthorization \t %s %s (%d)", x.User.FirstName, x.User.LastName, x.User.ID))
	case msg.C_AuthCheckedPhone:
		x := new(msg.AuthCheckedPhone)
		x.Unmarshal(envelope.Message)
		_Shell.Println(fmt.Sprintf("AuthCheckedPhone \t Registered:%t", x.Registered))
	case msg.C_AuthRecalled:
		x := new(msg.AuthRecalled)
		x.Unmarshal(envelope.Message)
		_Shell.Println(fmt.Sprintf("AuthRecalled \t ClientID:%d , Timestamp:%d", x.ClientID, x.Timestamp))
	case msg.C_AuthSentCode:
		x := new(msg.AuthSentCode)
		x.Unmarshal(envelope.Message)
		if strings.HasPrefix(x.Phone, "2374") {
			os.Remove("./_connection/phone")
			os.Remove("./_connection/phoneCodeHash")
			ioutil.WriteFile("./_connection/phone", []byte(x.Phone), 0666)
			ioutil.WriteFile("./_connection/phoneCodeHash", []byte(x.PhoneCodeHash), 0666)
		}
		_Shell.Println(fmt.Sprintf("AuthSentCode \t Phone:%s , Hash:%s", x.Phone, x.PhoneCodeHash))
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
		_Shell.Println("\r\n" + buf.String())
	case msg.C_ContactsMany:
		x := new(msg.ContactsMany)
		x.Unmarshal(envelope.Message)

		bufUsers := new(bytes.Buffer)
		tableUsers := tablewriter.NewWriter(bufUsers)
		tableUsers.SetCaption(true, "Users")
		tableUsers.SetHeader([]string{
			"FirstName", "LastName", "Username", "User ID", "AccessHash",
		})

		for _, u := range x.Users {
			tableUsers.Append([]string{
				u.FirstName,
				u.LastName,
				u.Username,
				fmt.Sprintf("%d", u.ID),
				fmt.Sprintf("%d", u.AccessHash),
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
			})
		}
		tableContacts.Render()
		_Shell.Println("\r\n" + bufUsers.String())
		_Shell.Println("\r\n" + bufContacts.String())
	case msg.C_MessagesDialogs:
		x := new(msg.MessagesDialogs)
		x.Unmarshal(envelope.Message)

		bufDialogs := new(bytes.Buffer)
		tableDialogs := tablewriter.NewWriter(bufDialogs)
		tableDialogs.SetHeader([]string{
			"PeerID", "PeerType", "Top Message ID", "Unread", "AccessHash", "MentionedCount",
		})

		for _, d := range x.Dialogs {
			tableDialogs.Append([]string{
				fmt.Sprintf("%d", d.PeerID),
				fmt.Sprintf("%d", d.PeerType),
				fmt.Sprintf("%d", d.TopMessageID),
				fmt.Sprintf("%d", d.UnreadCount),
				fmt.Sprintf("%d", d.AccessHash),
				fmt.Sprintf("%d", d.MentionedCount),
			})
		}
		tableDialogs.Render()
		bufUsers := new(bytes.Buffer)
		tableUsers := tablewriter.NewWriter(bufUsers)
		tableUsers.SetHeader([]string{
			"userID", "FirstName", "LastName", "Photo",
		})
		for _, x := range x.Users {
			tableUsers.Append([]string{
				fmt.Sprintf("%d", x.ID),
				fmt.Sprintf("%s", x.FirstName),
				fmt.Sprintf("%s", x.LastName),
				fmt.Sprintf("%d", len(x.Photo.String())),
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

		_Shell.Println("\r\n" + fmt.Sprintf("Total: %d", x.Count))
		_Shell.Println("\r\n" + bufDialogs.String())
		_Shell.Println("\r\n" + bufUsers.String())
		_Shell.Println("\r\n" + bufGroup.String())
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
		_Shell.Println("\r\n" + buf.String())
	case msg.C_MessagesSent:
		x := new(msg.MessagesSent)
		x.Unmarshal(envelope.Message)
		_Shell.Println(fmt.Sprintf("MessagesSent \t MsgID:%d , RandomID:%d", x.MessageID, x.RandomID))
	case msg.C_Bool:
		x := new(msg.Bool)
		x.Unmarshal(envelope.Message)
		_Shell.Println(fmt.Sprintf("Bool \t Res:%t", x.Result))
	case msg.C_Error:
		x := new(msg.Error)
		x.Unmarshal(envelope.Message)
		_Shell.Println(fmt.Sprintf("Error \t %s:%s", x.Code, x.Items))
	case msg.C_MessagesMany:

		x := new(msg.MessagesMany)
		x.Unmarshal(envelope.Message)
		bufMessages := new(bytes.Buffer)
		tableMessages := tablewriter.NewWriter(bufMessages)
		tableMessages.SetHeader([]string{
			"MsgID", "PeerID", "PeerType", "CreatedOn", "Flags", "Body", "Entities", "MeidaType",
		})

		for _, d := range x.Messages {
			tableMessages.Append([]string{
				fmt.Sprintf("%d", d.ID),
				fmt.Sprintf("%d", d.PeerID),
				fmt.Sprintf("%d", d.PeerType),
				fmt.Sprintf("%d", d.CreatedOn),
				fmt.Sprintf("%d", d.Flags),
				fmt.Sprintf("%v", string(d.Body)),
				fmt.Sprintf("%v", d.Entities),
				fmt.Sprintf("%s", d.MediaType.String()),
			})
		}
		tableMessages.Render()
		bufUsers := new(bytes.Buffer)
		tableUsers := tablewriter.NewWriter(bufUsers)
		tableUsers.SetHeader([]string{
			"userID", "FirstName", "LastName", "Photo",
		})
		for _, x := range x.Users {
			tableUsers.Append([]string{
				fmt.Sprintf("%d", x.ID),
				fmt.Sprintf("%s", x.FirstName),
				fmt.Sprintf("%s", x.LastName),
				fmt.Sprintf("%d", len(x.Photo.String())),
			})
		}
		tableUsers.Render()

		_Shell.Println(fmt.Sprintf("Total Message Count: %d", len(x.Messages)))
		_Shell.Println("\r\n" + bufMessages.String())
		_Shell.Println("\r\n" + bufUsers.String())
	case msg.C_UsersMany:

		x := new(msg.UsersMany)
		x.Unmarshal(envelope.Message)
		bufUsers := new(bytes.Buffer)
		tableUsers := tablewriter.NewWriter(bufUsers)
		tableUsers.SetHeader([]string{
			"userID", "FirstName", "LastName", "Photo",
		})
		for _, x := range x.Users {
			tableUsers.Append([]string{
				fmt.Sprintf("%d", x.ID),
				fmt.Sprintf("%s", x.FirstName),
				fmt.Sprintf("%d", len(x.Photo.String())),
				fmt.Sprintf("%s", x.LastName),
			})
		}
		tableUsers.Render()
		_Shell.Println("\r\n" + bufUsers.String())
	case msg.C_UpdateDifference:
		x := new(msg.UpdateDifference)
		x.Unmarshal(envelope.Message)

		_Shell.Println(fmt.Sprintf("Received UpdateDifference \t MaxID:%d \t MinID:%d \t UpdateCounts:%d", x.MaxUpdateID, x.MinUpdateID, len(x.Updates)))

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
		_Shell.Println("\r\n" + bufMsg.String())
	case msg.C_GroupFull:
		x := new(msg.GroupFull)
		err := x.Unmarshal(envelope.Message)
		if err != nil {
			_Log.Error("Failed to unmarshal", zap.Error(err))
			return
		}
		if x.Group != nil {
			_Shell.Println(fmt.Sprintf("GroupID : %d \t Title : %s \t Flags :%v", x.Group.ID, x.Group.Title, x.Group.Flags))
			if x.Group.Photo == nil {
				_Shell.Println("GroupPhoto is null")
			} else {
				_Shell.Println("GroupPhoto", zap.String("Big", x.Group.Photo.PhotoBig.String()), zap.String("Small", x.Group.Photo.PhotoSmall.String()))
			}

		} else {
			_Log.Error("x.Group is null")
		}
		if x.NotifySettings != nil {
			_Shell.Println(fmt.Sprintf("NotifySettings Sound: %s \t Mute : %d \t Flag : %d", x.NotifySettings.Sound, x.NotifySettings.MuteUntil, x.NotifySettings.Flags))
		} else {
			_Log.Error("x.NotifySettings is null")
		}
		if x.Participants != nil {
			_Shell.Println(fmt.Sprintf("Participants Count : %d ", len(x.Participants)))

			bufUsers := new(bytes.Buffer)
			tableUsers := tablewriter.NewWriter(bufUsers)
			tableUsers.SetHeader([]string{
				"userID", "FirstName", "LastName", "AccessHash", "Username", "Photo",
			})
			for _, x := range x.Participants {
				tableUsers.Append([]string{
					fmt.Sprintf("%d", x.UserID),
					fmt.Sprintf("%s", x.FirstName),
					fmt.Sprintf("%s", x.LastName),
					fmt.Sprintf("%d", x.AccessHash),
					fmt.Sprintf("%s", x.Username),
					fmt.Sprintf("%d", len(x.Photo.String())),
				})
			}
			tableUsers.Render()
			_Shell.Println("\r\n" + bufUsers.String())

		} else {
			_Log.Error("x.Participants is null")
		}
	case msg.C_InputUser:
		x := new(msg.InputUser)
		x.Unmarshal(envelope.Message)
		bufUsers := new(bytes.Buffer)
		tableUsers := tablewriter.NewWriter(bufUsers)
		tableUsers.SetHeader([]string{
			"userID", "AccessHash",
		})
		tableUsers.Append([]string{
			fmt.Sprintf("%d", x.UserID),
			fmt.Sprintf("%d", x.AccessHash),
		})
		tableUsers.Render()
		_Shell.Println("\r\n" + bufUsers.String())
	case msg.C_SystemServerTime:
		x := new(msg.SystemServerTime)
		x.Unmarshal(envelope.Message)
		serverTime := x.Timestamp
		clientTime := time.Now().Unix()
		delta := serverTime - clientTime
		_Shell.Println(fmt.Sprintf("ServerTime : %d \t ClientTime : %d \t Delta: %d", serverTime, clientTime, delta))
	case msg.C_UpdateState:
		x := new(msg.UpdateState)
		x.Unmarshal(envelope.Message)
		_Shell.Println("\r\n" + x.String())

	case msg.C_BotsMany:
		x := new(msg.BotsMany)
		x.Unmarshal(envelope.Message)
		_Shell.Println("River bots info for user(", _SDK.ConnInfo.UserID, "): ")
		for _, bot := range x.Bots {
			_Shell.Println("Bot {", bot.Bot.Username, "} info:")
			_Shell.Println("ID:", bot.Bot.ID)
			_Shell.Println("Name:", bot.Bot.Name)
			_Shell.Println("OwnerID:", bot.UserID)
			_Shell.Println("Description:", bot.Description)
			_Shell.Println("Commands:")
			for _, cmd := range bot.BotCommands {
				_Shell.Println(cmd.Command, "-", cmd.Description)
			}
		}

	default:
		constructorName, _ := msg.ConstructorNames[envelope.Constructor]
		_Shell.Println("DEFAULT",
			zap.String("ConstructorName", constructorName),
			zap.Int64("Constructor", envelope.Constructor),
		)
	}
}
