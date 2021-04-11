package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"git.ronaksoft.com/river/msg/go/msg"
	"gopkg.in/abiosoft/ishell.v2"
)

func fnGetString(c *ishell.Context, prompt string) string {
	c.Print(prompt, ": ")
	x := c.ReadLine()
	return x
}

func fnGetPhone(c *ishell.Context) string {
	c.Print("Phone: ")
	phone := c.ReadLine()
	return phone
}

func fnGetPhoneCode(c *ishell.Context) string {
	c.Print("Phone Code: ")
	code := c.ReadLine()
	return code
}

func fnGetPhoneCodeHash(c *ishell.Context) string {
	c.Print("Phone Code Hash: ")
	hash := c.ReadLine()
	return hash
}

func fnGetFirstName(c *ishell.Context) string {
	c.Print("First Name: ")
	fName := c.ReadLine()
	return fName
}

func fnGetLastName(c *ishell.Context) string {
	c.Print("Last Name: ")
	lName := c.ReadLine()
	return lName
}

func fnGetPeerID(c *ishell.Context) int64 {
	var peerID int64
	for {
		c.Print("Peer ID: ")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil {
			peerID = id
			break
		} else {
			c.Println(err.Error())
		}
	}
	return peerID
}

func fnGetBotID(c *ishell.Context) int64 {
	var botID int64
	for {
		c.Print("Bot ID: ")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil {
			botID = id
			break
		} else {
			c.Println(err.Error())
		}
	}
	return botID
}

func fnGetPeerType(c *ishell.Context) msg.PeerType {
	var peerType msg.PeerType

	for {
		c.Print("Peer Type: ")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil {
			switch id {
			case 0:
				peerType = msg.PeerType_PeerSelf
			case 1:
				peerType = msg.PeerType_PeerUser
			case 2:
				peerType = msg.PeerType_PeerGroup
			case 3:
				peerType = msg.PeerType_PeerChannel
			default:
				c.Println("Invalid peerType (0:self,1:user,2:group,3:channel)")
			}
			break
		} else {
			c.Println(err.Error())
		}
	}

	return peerType
}

func fnGetAccessHash(c *ishell.Context) uint64 {
	var accessHash uint64
	for {
		c.Print("Access Hash: ")
		hash, err := strconv.ParseUint(c.ReadLine(), 10, 64)
		if err == nil {
			accessHash = hash
			break
		} else {
			c.Println(err.Error())
		}
	}
	return accessHash
}

func fnGetTries(c *ishell.Context) int {
	var count int
	for {
		c.Print("Tries : ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			count = int(tmp)
			break
		} else {
			c.Println(err.Error())
		}
	}
	return count
}

func fnGetInterval(c *ishell.Context) time.Duration {
	var interval time.Duration
	for {
		c.Print("Interval: ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			interval = time.Duration(tmp) * time.Millisecond
			break
		} else {
			c.Println(err.Error())
		}
	}
	return interval
}

func fnGetBody(c *ishell.Context) string {
	c.Print("Body: ")
	body := c.ReadLine()
	return body
}

func fnGetPassword(c *ishell.Context) []byte {
	c.Print("Password: ")
	body := c.ReadLine()
	return []byte(body)
}

func fnGetMaxID(c *ishell.Context) int64 {
	var maxID int64
	for {
		c.Print("Max ID: ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil {
			maxID = tmp
			break
		} else {
			c.Println(err.Error())
		}
	}
	return maxID
}

func fnGetMinID(c *ishell.Context) int64 {
	var minID int64
	for {
		c.Print("Min ID: ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil {
			minID = tmp
			break
		} else {
			c.Println(err.Error())
		}
	}
	return minID
}

func fnGetLabelID(c *ishell.Context) int32 {
	var labelID int32
	for {
		c.Print("LabelID: ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			labelID = int32(tmp)
			break
		} else {
			c.Println(err.Error())
		}
	}
	return labelID
}

func fnGetLimit(c *ishell.Context) int32 {
	var limit int32
	for {
		c.Print("Limit: ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			limit = int32(tmp)
			break
		} else {
			c.Println(err.Error())
		}
	}
	return limit
}

func fnGetTypingAction(c *ishell.Context) msg.TypingAction {
	var action msg.TypingAction
	for {
		c.Print("Action (0:Typing, 4:Cancel): ")
		actionID, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			action = msg.TypingAction(actionID)
			break
		} else {
			c.Println(err.Error())
		}
	}
	return action
}

func fnGetMessageIDs(c *ishell.Context) []int64 {
	messagesIDs := make([]int64, 0)
	for {

		c.Print(len(messagesIDs), "Enter none numeric character to break\r\n")
		c.Print(len(messagesIDs), "MessageID: ")
		msgID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err != nil {
			break
		} else {
			messagesIDs = append(messagesIDs, msgID)
		}
	}
	return messagesIDs
}

func fnGetMessageID(c *ishell.Context) int64 {
	messageID := int64(0)
	for {
		c.Print("MessageID: ")
		msgID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err != nil {
			break
		} else {
			messageID = msgID
			break
		}
	}
	return messageID
}

func fnGetFromUpdateID(c *ishell.Context) int64 {
	var updateID int64
	for {
		c.Print("From UpdateID: ")
		fromUpdateID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil {
			updateID = fromUpdateID
			break
		} else {
			c.Println(err.Error())
		}
	}
	return updateID
}

func fnGetInputUser(c *ishell.Context) []*msg.InputUser {
	users := make([]*msg.InputUser, 0)
	for {
		c.Print("Enter none numeric character to break\r\n")

		c.Print(len(users), "User ID: ")
		userID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err != nil {
			break
		}

		c.Print(len(users), "Access Hash: ")
		accessHash, err := strconv.ParseUint(c.ReadLine(), 10, 64)
		if err != nil {
			break
		}

		u := new(msg.InputUser)
		u.UserID = userID
		u.AccessHash = accessHash
		users = append(users, u)
	}
	return users
}

func fnGetUsername(c *ishell.Context) string {
	c.Print("Username: ")
	uname := c.ReadLine()
	return uname
}

func fnGetQuery(c *ishell.Context) string {
	c.Print("Query: ")
	uname := c.ReadLine()
	return uname
}

func fnGetResultID(c *ishell.Context) string {
	c.Print("ResultID: ")
	uname := c.ReadLine()
	return uname
}

func fnGetQueryID(c *ishell.Context) int64 {
	var queryID int64
	for {
		c.Print("QueryID: ")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil {
			queryID = id
			break
		} else {
			c.Println(err.Error())
		}
	}
	return queryID
}

func fnGetTitle(c *ishell.Context) string {
	c.Print("Title: ")
	title := c.ReadLine()
	return title
}

func fnGetGroupID(c *ishell.Context) int64 {
	var groupID int64
	for {
		c.Print("Group ID: ")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil {
			groupID = id
			break
		} else {
			c.Println(err.Error())
		}
	}
	return groupID
}

func fnGetForwardLimit(c *ishell.Context) int32 {
	var fwdLimit int32
	for {
		c.Print("Forward Limit: ")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			fwdLimit = int32(id)
			break
		} else {
			c.Println(err.Error())
		}
	}
	return fwdLimit
}

func fnGetRevoke(c *ishell.Context) bool {
	revoke := false
	for {
		c.Print("Revoke : (0 = false , >=1 : true)")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			revoke = id > 0
			break
		} else {
			c.Println(err.Error())
		}
	}
	return revoke
}

func fnGetSilence(c *ishell.Context) bool {
	silence := false
	for {
		c.Print("Silence : (0 = false , >=1 : true)")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			silence = id > 0
			break
		} else {
			c.Println(err.Error())
		}
	}
	return silence
}

func fnGetDelete(c *ishell.Context) bool {
	del := false
	for {
		c.Print("Delete : (0 = false , >=1 : true)")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			del = id > 0
			break
		} else {
			c.Println(err.Error())
		}
	}
	return del
}

func fnGetAdmin(c *ishell.Context) bool {
	del := false
	for {
		c.Print("Admin : (0 = false , >=1 : true)")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			del = id > 0
			break
		} else {
			c.Println(err.Error())
		}
	}
	return del
}

func fnGetAdminEnabled(c *ishell.Context) bool {
	del := false
	for {
		c.Print("Admin Enabled : (0 = false , >=1 : true)")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			del = id > 0
			break
		} else {
			c.Println(err.Error())
		}
	}
	return del
}

func fnGetEntities(c *ishell.Context) []*msg.MessageEntity {
	entities := make([]*msg.MessageEntity, 0)
	for {
		c.Print("Enter none numeric character to break\r\n")
		var entityType msg.MessageEntityType
		var offset int32
		var length int32
		var userID int64
		for {
			c.Print(len(entities), "Type: (0:Bold, 1:Italic, 2:Mention,3:Url, 4:Email ,5:Hashtag)")
			typeID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
			if err == nil && typeID < 6 {
				entityType = msg.MessageEntityType(typeID)
				break
			} else {
				return entities
			}
		}

		for {
			c.Print(len(entities), "Offset: ")
			tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
			if err == nil {
				offset = int32(tmp)
				break
			} else {
				return entities
			}
		}

		for {
			c.Print(len(entities), "Length: ")
			tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
			if err == nil {
				length = int32(tmp)
				break
			} else {
				return entities
			}
		}

		for {
			c.Print(len(entities), "userID: ")
			tmp, err := strconv.ParseInt(c.ReadLine(), 10, 64)
			if err == nil {
				userID = tmp
				break
			} else {
				return entities
			}
		}

		e := &msg.MessageEntity{
			Length: length,
			Offset: offset,
			Type:   entityType,
			UserID: userID,
		}
		entities = append(entities, e)
	}
}

func fnGetTokenType(c *ishell.Context) msg.PushTokenProvider {
	var tokenType msg.PushTokenProvider
	for {
		c.Print("TokenType (Firebase = 0 , APN = 1) : ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			tokenType = msg.PushTokenProvider(tmp)
			break
		}
	}
	return tokenType
}

func fnGetToken(c *ishell.Context) string {
	c.Print("Token: ")
	token := c.ReadLine()
	return token
}

func fnGetProvider(c *ishell.Context) string {
	c.Print("Provider (ap, nested, google, apple): ")
	token := c.ReadLine()
	return token
}

func fnGetDeviceModel(c *ishell.Context) string {
	c.Print("Model(ios | android) : ")
	model := c.ReadLine()
	return model
}

func fnGetSysytemVersion(c *ishell.Context) string {
	c.Print("Sysytem Version : ")
	version := c.ReadLine()
	return version
}

func fnGetAppVersion(c *ishell.Context) string {
	c.Print("App Version : ")
	version := c.ReadLine()
	return version
}

func fnGetLangCode(c *ishell.Context) string {
	c.Print("Language Code : ")
	code := c.ReadLine()
	return code
}

func fnGetClientID(c *ishell.Context) string {
	c.Print("Client ID : ")
	code := c.ReadLine()
	return code
}

func fnGetLabelName(c *ishell.Context) string {
	c.Print("Label Name: ")
	name := c.ReadLine()
	return name
}

func fnGetLabelColour(c *ishell.Context) string {
	c.Print("Label Colour: ")
	name := c.ReadLine()
	return name
}

func fnGetFilePath(c *ishell.Context) string {
	c.Print("File Path: ")
	name := c.ReadLine()
	return name
}

func fnGetThumbFilePath(c *ishell.Context) string {
	c.Print("Thumbnail File Path: ")
	name := c.ReadLine()
	return name
}

func fnGetReplyTo(c *ishell.Context) int64 {
	var replyTo int64
	for {
		c.Print("Reply To: ")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil {
			replyTo = id
			break
		} else {
			c.Println(err.Error())
		}
	}
	return replyTo
}

func fnGetPeer(c *ishell.Context) *msg.InputPeer {
	options := make([]string, 0, len(MyDialogs))
	for _, d := range MyDialogs {
		switch d.PeerType {
		case 1:
			options = append(options, fmt.Sprintf("%s %s (%d)", MyUsers[d.PeerID].FirstName, MyUsers[d.PeerID].LastName, d.TopMessageID))
		case 2:
			options = append(options, fmt.Sprintf("%s (%d)", MyGroups[d.PeerID].Title, d.TopMessageID))
		}

	}
	idx := c.MultiChoice(options, "Please Select Your Dialog:")
	return &msg.InputPeer{
		ID:         MyDialogs[idx].PeerID,
		Type:       msg.PeerType(MyDialogs[idx].PeerType),
		AccessHash: MyDialogs[idx].AccessHash,
	}
}

func fnGetUser(c *ishell.Context) *msg.InputUser {
	options := make([]string, 0, len(MyUsers))
	optionsUser := make([]*msg.User, 0, len(MyUsers))
	for _, u := range MyUsers {
		optionsUser = append(optionsUser, MyUsers[u.ID])
	}
	sort.Slice(optionsUser, func(i, j int) bool {
		return strings.Compare(optionsUser[i].FirstName, optionsUser[j].FirstName) < 0
	})

	for idx := range optionsUser {
		options = append(options, fmt.Sprintf("%s %s", optionsUser[idx].FirstName, optionsUser[idx].LastName))
	}

	idx := c.MultiChoice(options, "Please Select Your User:")
	return &msg.InputUser{
		UserID:     optionsUser[idx].ID,
		AccessHash: optionsUser[idx].AccessHash,
	}
}

func fnGetBot(c *ishell.Context) *msg.InputUser {
	options := make([]string, 0, len(MyUsers))
	optionsUser := make([]*msg.User, 0, len(MyUsers))
	for _, u := range MyUsers {
		if u.IsBot {
			options = append(options, fmt.Sprintf("%s %s", MyUsers[u.ID].FirstName, MyUsers[u.ID].LastName))
			optionsUser = append(optionsUser, MyUsers[u.ID])
		}
	}
	idx := c.MultiChoice(options, "Please Select Your Bot:")
	return &msg.InputUser{
		UserID:     optionsUser[idx].ID,
		AccessHash: optionsUser[idx].AccessHash,
	}
}

func fnGetTopPeerCat(c *ishell.Context) msg.TopPeerCategory {
	options := []string{
		msg.TopPeerCategory_Users.String(),
		msg.TopPeerCategory_Groups.String(),
		msg.TopPeerCategory_Forwards.String(),
		msg.TopPeerCategory_BotsMessage.String(),
		msg.TopPeerCategory_BotsInline.String(),
	}
	idx := c.MultiChoice(options, "Select the category:")
	return msg.TopPeerCategory(idx)

}

func fnGetUpdateNewMessageHexString(c *ishell.Context) string {
	c.Print("Enter UpdateNewMessage Hex String:")
	title := c.ReadLine()

	if title[:2] == "0x" {
		title = title[2:]
	}
	return title
}

func fnGetMime(c *ishell.Context) string {
	c.Print("MIME Type:")
	mime := c.ReadLine()
	return mime
}

func fnGetMediaType(c *ishell.Context) msg.ClientMediaType {
	mediaType := msg.ClientMediaType_ClientMediaNone
	for {
		c.Print("Media Type : (All=0, File= 1, Media= 2, Voice= 3, Audio= 4)")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil && id < 5 {
			mediaType = msg.ClientMediaType(id)
			break
		} else {
			c.Println("entered value is invalid ")
		}
	}
	return mediaType
}

func fnGetMediaCat(c *ishell.Context) msg.MediaCategory {
	options := make([]string, len(msg.MediaCategory_value))
	for n, v := range msg.MediaCategory_value {
		options[v] = n
	}

	idx := c.MultiChoice(options, "Please Select Category:")
	return msg.MediaCategory(idx)
}

func fnGetFileID(c *ishell.Context) int64 {
	var res int64
	for {
		c.Print("FileID : ")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil {
			res = id
			break
		} else {
			c.Println(err.Error())
		}
	}
	return res
}

func fnGetOffset(c *ishell.Context) int32 {
	var res int32
	for {
		c.Print("Offset : ")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			res = int32(id)
			break
		} else {
			c.Println(err.Error())
		}
	}
	return res
}

func fnGetTeamID(c *ishell.Context) int64 {
	var res int64
	for {
		c.Print("TeamID : ")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil {
			res = id
			break
		} else {
			c.Println(err.Error())
		}
	}
	return res
}

func fnGetUserID(c *ishell.Context) int64 {
	var res int64
	for {
		c.Print("UserID : ")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil {
			res = id
			break
		} else {
			c.Println(err.Error())
		}
	}
	return res
}
