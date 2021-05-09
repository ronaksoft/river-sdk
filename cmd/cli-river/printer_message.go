package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/registry"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"git.ronaksoft.com/river/msg/go/msg"
	"github.com/olekukonko/tablewriter"
)

var (
	MyDialogs []*msg.Dialog
	MyUsers   = map[int64]*msg.User{}
	MyGroups  = map[int64]*msg.Group{}
	Printers  = map[int64]domain.MessageApplier{}
)

func init() {
	Printers = map[int64]domain.MessageApplier{
		msg.C_AccountPassword:      printAccountPassword,
		msg.C_AuthAuthorization:    printAuthAuthorization,
		msg.C_AuthCheckedPhone:     printAuthCheckedPhone,
		msg.C_AuthRecalled:         printAuthRecalled,
		msg.C_AuthSentCode:         printAuthSentCode,
		msg.C_ClientPendingMessage: printClientPendingMessage,
		msg.C_ClientSearchResult:   printClientSearchResult,
		msg.C_ContactsImported:     printContactsImported,
		msg.C_ContactsMany:         printContactsMany,
		msg.C_Dialog:               printDialog,
		msg.C_GroupFull:            printGroupFull,
		msg.C_LabelItems:           printLabelItems,
		msg.C_LabelsMany:           printLabelsMany,
		msg.C_MessagesDialogs:      printMessagesDialogs,
		msg.C_MessagesMany:         printMessagesMany,
		msg.C_MessagesSent:         printMessagesSent,
		msg.C_SystemConfig:         printSystemConfig,
		msg.C_TeamsMany:            printTeamsMany,
		msg.C_UpdateDifference:     printUpdateDifference,
		msg.C_UsersMany:            printUsersMany,
		rony.C_Error:               printError,
	}
}

func MessagePrinter(envelope *rony.MessageEnvelope) {
	p := Printers[envelope.Constructor]
	if p != nil {
		p(envelope)
		return
	}

	switch envelope.Constructor {
	case msg.C_Bool:
		x := new(msg.Bool)
		x.Unmarshal(envelope.Message)
		_Shell.Println(fmt.Sprintf("Bool \t Res:%t", x.Result))
	case msg.C_UpdateNewMessage:
		x := new(msg.UpdateNewMessage)
		x.Unmarshal(envelope.Message)

		bufMsg := new(bytes.Buffer)
		tableMsg := tablewriter.NewWriter(bufMsg)
		tableMsg.SetHeader([]string{
			"GetUpdateID", "AccessHash", "Sender", "Message.ID", "Message.Body",
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
	case msg.C_BotCommandsMany:
		x := new(msg.BotCommandsMany)
		x.Unmarshal(envelope.Message)
		_Shell.Println("Available commands: ")
		for _, cmd := range x.Commands {
			_Shell.Println(cmd.Command, "-", cmd.Description)
		}
	case msg.C_ContactsTopPeers:
		x := &msg.ContactsTopPeers{}
		x.Unmarshal(envelope.Message)
		for _, tp := range x.Peers {
			_Shell.Println(tp.Peer.ID, tp.Peer.Type, tp.Rate)
		}
	case msg.C_WallPapersMany:
		x := &msg.WallPapersMany{}
		x.Unmarshal(envelope.Message)
		for _, wp := range x.WallPapers {
			_Shell.Println(wp.ID, wp.AccessHash, wp.Creator, wp.Document.ID, wp.Document.AccessHash)
		}
	case msg.C_ClientFilesMany:
		x := &msg.ClientFilesMany{}
		x.Unmarshal(envelope.Message)
		_Shell.Println(x.Total, len(x.Gifs))
		for _, g := range x.Gifs {
			_Shell.Println(g.FileID, g.AccessHash)
		}
	case msg.C_SavedGifs:
		x := &msg.SavedGifs{}
		x.Unmarshal(envelope.Message)
		_Shell.Println(x.Hash, x.NotModified, len(x.Docs))
		for _, d := range x.Docs {
			fmt.Println(d.Caption, d.Doc.ID, d.Doc.Attributes)
		}
	case msg.C_TeamMembers:
		x := &msg.TeamMembers{}
		_ = x.Unmarshal(envelope.Message)
		for _, m := range x.Members {
			_Shell.Println(m.Admin, m.UserID, m.User.Username, m.User.FirstName, m.User.LastName)
		}
	case msg.C_BotResults:
		x := &msg.BotResults{}
		_ = x.Unmarshal(envelope.Message)

		buf := new(bytes.Buffer)
		table := tablewriter.NewWriter(buf)
		table.SetHeader([]string{
			"ID", "Type", "Title", "Message (Size)",
		})
		for _, r := range x.Results {
			table.Append([]string{
				r.ID,
				r.Type.String(),
				r.Title,
				fmt.Sprintf("%d", proto.Size(r.Message)),
			})
		}
		_Shell.Println("QueryID:", x.QueryID)
		_Shell.Println("NextOffset:", x.NextOffset)
		table.Render()
		_Shell.Println(buf)
	default:
		constructorName := registry.ConstructorName(envelope.Constructor)
		_Shell.Println("DEFAULT", constructorName, len(envelope.Message))
	}
}

func printTeamsMany(envelope *rony.MessageEnvelope) {
	x := &msg.TeamsMany{}
	x.Unmarshal(envelope.Message)
	for _, t := range x.Teams {
		_Shell.Println(t.ID, t.AccessHash, t.Name, t.CreatorID)
	}
}
func printMessagesMany(envelope *rony.MessageEnvelope) {
	x := new(msg.MessagesMany)
	x.Unmarshal(envelope.Message)
	bufMessages := new(bytes.Buffer)
	tableMessages := tablewriter.NewWriter(bufMessages)
	tableMessages.SetHeader([]string{
		"MsgID", "PeerID", "PeerType", "CreatedOn", "Flags", "Body", "Entities", "MeidaType", "FileID", "AccessHash",
	})

	for _, d := range x.Messages {
		var docID int64
		var accessHash uint64
		if d.MediaType == msg.MediaType_MediaTypeDocument {
			xx := &msg.MediaDocument{}
			xx.Unmarshal(d.Media)
			docID = xx.Doc.ID
			accessHash = xx.Doc.AccessHash
		} else {
			docID = 0
			accessHash = 0
		}
		tableMessages.Append([]string{
			fmt.Sprintf("%d", d.ID),
			fmt.Sprintf("%d", d.PeerID),
			fmt.Sprintf("%d", d.PeerType),
			fmt.Sprintf("%d", d.CreatedOn),
			fmt.Sprintf("%d", d.Flags),
			fmt.Sprintf("%v", d.Body),
			fmt.Sprintf("%v", d.Entities),
			fmt.Sprintf("%s", d.MediaType.String()),
			fmt.Sprintf("%d", docID),
			fmt.Sprintf("%d", accessHash),
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
}
func printUsersMany(envelope *rony.MessageEnvelope) {
	x := new(msg.UsersMany)
	x.Unmarshal(envelope.Message)
	bufUsers := new(bytes.Buffer)
	tableUsers := tablewriter.NewWriter(bufUsers)
	tableUsers.SetHeader([]string{
		"userID", "FirstName", "LastName", "Username", "Photo", "LastSeen",
	})
	for _, x := range x.Users {
		tableUsers.Append([]string{
			fmt.Sprintf("%d", x.ID),
			fmt.Sprintf("%s", x.FirstName),
			fmt.Sprintf("%s", x.LastName),
			x.Username,
			fmt.Sprintf("%d", x.Photo.PhotoID),
			fmt.Sprintf("%s", time.Unix(x.LastSeen, 0).Format(time.RFC822)),
		})
	}
	tableUsers.Render()
	_Shell.Println("\r\n" + bufUsers.String())
}
func printUpdateDifference(envelope *rony.MessageEnvelope) {
	x := new(msg.UpdateDifference)
	x.Unmarshal(envelope.Message)

	_Shell.Println(fmt.Sprintf("Received UpdateDifference \t MaxID:%d \t MinID:%d \t UpdateCounts:%d", x.MaxUpdateID, x.MinUpdateID, len(x.Updates)))

	for _, v := range x.Updates {
		if v.Constructor == msg.C_UpdateNewMessage {
			msg := new(rony.MessageEnvelope)
			msg.Constructor = v.Constructor
			msg.Message = v.Update

			MessagePrinter(msg)
		}

	}
}
func printGroupFull(envelope *rony.MessageEnvelope) {
	x := new(msg.GroupFull)
	err := x.Unmarshal(envelope.Message)
	if err != nil {
		_Shell.Println("Failed to unmarshal", zap.Error(err))
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
		_Shell.Println("x.Group is null")
	}
	if x.NotifySettings != nil {
		_Shell.Println(fmt.Sprintf("NotifySettings Sound: %s \t Mute : %d \t Flag : %d", x.NotifySettings.Sound, x.NotifySettings.MuteUntil, x.NotifySettings.Flags))
	} else {
		_Shell.Println("x.NotifySettings is null")
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
		_Shell.Println("x.Participants is null")
	}
}
func printLabelsMany(envelope *rony.MessageEnvelope) {
	x := &msg.LabelsMany{}
	x.Unmarshal(envelope.Message)
	for _, l := range x.Labels {
		_Shell.Println(l.ID, l.Count, l.Name)
	}
}
func printLabelItems(envelope *rony.MessageEnvelope) {
	x := &msg.LabelItems{}
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
			fmt.Sprintf("%v", d.Body),
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
}
func printContactsImported(envelope *rony.MessageEnvelope) {
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
}
func printContactsMany(envelope *rony.MessageEnvelope) {
	x := new(msg.ContactsMany)
	x.Unmarshal(envelope.Message)

	bufUsers := new(bytes.Buffer)
	tableUsers := tablewriter.NewWriter(bufUsers)
	tableUsers.SetCaption(true, "Users")
	tableUsers.SetHeader([]string{
		"FirstName", "LastName", "Username", "User ID", "AccessHash", "Phone",
	})

	for _, u := range x.ContactUsers {
		tableUsers.Append([]string{
			u.FirstName,
			u.LastName,
			u.Username,
			fmt.Sprintf("%d", u.ID),
			fmt.Sprintf("%d", u.AccessHash),
			u.Phone,
		})
	}
	tableUsers.Render()
	_Shell.Println("\r\n" + bufUsers.String())

}
func printMessagesDialogs(envelope *rony.MessageEnvelope) {
	x := new(msg.MessagesDialogs)
	x.Unmarshal(envelope.Message)

	bufDialogs := new(bytes.Buffer)
	tableDialogs := tablewriter.NewWriter(bufDialogs)
	tableDialogs.SetHeader([]string{
		"PeerID", "PeerType", "Top Message ID", "Unread", "AccessHash", "MentionedCount",
	})
	MyDialogs = append(MyDialogs[:0], x.Dialogs...)
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
	for _, x := range x.Users {
		MyUsers[x.ID] = x
	}
	for _, x := range x.Groups {
		MyGroups[x.ID] = x
	}

	_Shell.Println("\r\n" + fmt.Sprintf("Total: %d", x.Count))
	_Shell.Println("\r\n" + bufDialogs.String())
}
func printDialog(envelope *rony.MessageEnvelope) {
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
}
func printMessagesSent(envelope *rony.MessageEnvelope) {
	x := new(msg.MessagesSent)
	x.Unmarshal(envelope.Message)
	_Shell.Println(fmt.Sprintf("MessagesSent \t MsgID:%d , RandomID:%d", x.MessageID, x.RandomID))
}
func printError(envelope *rony.MessageEnvelope) {
	x := new(rony.Error)
	_ = x.Unmarshal(envelope.Message)
	_Shell.Println(fmt.Sprintf("Error \t %s:%s (%s)", x.Code, x.Items, x.Description))
}
func printSystemConfig(envelope *rony.MessageEnvelope) {
	x := &msg.SystemConfig{}
	x.Unmarshal(envelope.Message)
	_Shell.Println("Reactions", x.Reactions)
}
func printClientPendingMessage(envelope *rony.MessageEnvelope) {
	x := &msg.ClientPendingMessage{}
	x.Unmarshal(envelope.Message)
	_Shell.Println(fmt.Sprintf("ClientPendingMessage ID:%d, FileID:%d, FileUploadID:%s", x.ID, x.FileID, x.FileUploadID))
}
func printAuthAuthorization(envelope *rony.MessageEnvelope) {
	x := new(msg.AuthAuthorization)
	x.Unmarshal(envelope.Message)
	_Shell.Println(fmt.Sprintf("AuthAuthorization \t %s %s (%d)", x.User.FirstName, x.User.LastName, x.User.ID))
}
func printAuthCheckedPhone(envelope *rony.MessageEnvelope) {
	x := new(msg.AuthCheckedPhone)
	x.Unmarshal(envelope.Message)
	_Shell.Println(fmt.Sprintf("AuthCheckedPhone \t Registered:%t", x.Registered))
}
func printAuthRecalled(envelope *rony.MessageEnvelope) {
	x := new(msg.AuthRecalled)
	x.Unmarshal(envelope.Message)
	_Shell.Println(fmt.Sprintf("AuthRecalled \t ClientID:%d , Timestamp:%d", x.ClientID, x.Timestamp))
}
func printAuthSentCode(envelope *rony.MessageEnvelope) {
	x := new(msg.AuthSentCode)
	x.Unmarshal(envelope.Message)
	if strings.HasPrefix(x.Phone, "2374") {
		os.Remove("./_phone")
		os.Remove("./_phoneCodeHash")
		ioutil.WriteFile("./_phone", []byte(x.Phone), 0666)
		ioutil.WriteFile("./_phoneCodeHash", []byte(x.PhoneCodeHash), 0666)
	}
	_Shell.Println(fmt.Sprintf("AuthSentCode \t Phone:%s , Hash:%s", x.Phone, x.PhoneCodeHash))
}
func printAccountPassword(envelope *rony.MessageEnvelope) {
	x := &msg.AccountPassword{}
	x.Unmarshal(envelope.Message)
	os.Remove("./_password")
	ioutil.WriteFile("./_password", envelope.Message, 0666)
	_Shell.Println("SrpB:", hex.EncodeToString(x.SrpB))
	_Shell.Println("SrpID:", x.SrpID)
	_Shell.Println("Hint:", x.Hint)
}
func printClientSearchResult(envelope *rony.MessageEnvelope) {
	x := &msg.ClientSearchResult{}
	x.Unmarshal(envelope.Message)
	bufUsers := new(bytes.Buffer)
	tableUsers := tablewriter.NewWriter(bufUsers)
	tableUsers.SetHeader([]string{
		"userID", "FirstName", "LastName", "Username", "LastSeen",
	})
	for _, x := range x.Users {
		tableUsers.Append([]string{
			fmt.Sprintf("%d", x.ID),
			fmt.Sprintf("%s", x.FirstName),
			fmt.Sprintf("%s", x.LastName),
			x.Username,
			fmt.Sprintf("%s", time.Unix(x.LastSeen, 0).Format(time.RFC822)),
		})
	}
	tableUsers.Render()
	_Shell.Println("\r\n" + bufUsers.String())
}
