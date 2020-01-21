package main

import (
	"crypto/rand"
	"strconv"
	"time"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/chat"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

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

func fnGetSearchPhrase(c *ishell.Context) string {
	c.Print("Search Phrase:")
	phrase := c.ReadLine()
	return phrase
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

func fnGetPeerType(c *ishell.Context) msg.PeerType {
	var peerType msg.PeerType

	for {
		c.Print("Peer Type: ")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil {
			switch id {
			case 0:
				peerType = msg.PeerSelf
			case 1:
				peerType = msg.PeerUser
			case 2:
				peerType = msg.PeerGroup
			case 3:
				peerType = msg.PeerChannel
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
			accessHash = uint64(hash)
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

func fnGetRequestID(c *ishell.Context) int64 {
	var requestID int64
	for {
		c.Print("RequestID : ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			requestID = tmp
			break
		} else {
			c.Println(err.Error())
		}
	}
	return requestID
}

func fnGetBody(c *ishell.Context) string {
	c.Print("Body: ")
	body := c.ReadLine()
	return body
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

func fnGetFileName(c *ishell.Context) string {
	c.Print("File Name: ")
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

func fnGetWidth(c *ishell.Context) uint32 {
	var res uint32
	for {
		c.Print("Width : ")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			res = uint32(id)
			break
		} else {
			c.Println(err.Error())
		}
	}
	return res
}
func fnGetHeight(c *ishell.Context) uint32 {
	var res uint32
	for {
		c.Print("Heigth : ")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			res = uint32(id)
			break
		} else {
			c.Println(err.Error())
		}
	}
	return res
}

func fnGetVoice(c *ishell.Context) bool {
	res := false
	for {
		c.Print("IsVoice : (0 = false , >=1 : true)")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			res = id > 0
			break
		} else {
			c.Println(err.Error())
		}
	}
	return res
}

func fnGetRound(c *ishell.Context) bool {
	res := false
	for {
		c.Print("IsRound : (0 = false , >=1 : true)")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			res = id > 0
			break
		} else {
			c.Println(err.Error())
		}
	}
	return res
}

func fnGetPerformer(c *ishell.Context) string {
	c.Print("Performer: ")
	title := c.ReadLine()
	return title
}

func fnGetWaveForm(c *ishell.Context) []byte {
	res := make([]byte, 100)
	rand.Read(res)
	return res
}

func fnGetInputMediaType(c *ishell.Context) msg.InputMediaType {
	mediaType := msg.InputMediaTypeEmpty
	for {
		c.Print("InputMediaType : (UploadedPhoto=1, UploadedDocument= 4)")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil && id == 1 || id == 4 {
			mediaType = msg.InputMediaType(id)
			break
		} else {
			c.Println("entered value is invalid ")
		}
	}
	return mediaType
}

func fnGetAttributes(c *ishell.Context) []*msg.DocumentAttribute {
	result := make([]*msg.DocumentAttribute, 0)

	for {
		c.Println("Enter 0 zero to break Attribute loop")
		attrType := getAttributeType(c)
		if attrType == msg.AttributeTypeNone {
			break
		}
		switch attrType {

		case msg.AttributeTypeAudio:
			result = append(result, getAudioAttribute(c))
		case msg.AttributeTypeVideo:
			result = append(result, getVideoAttribute(c))
		case msg.AttributeTypePhoto:
			result = append(result, getPhotoAttribute(c))
		case msg.AttributeTypeFile:
			result = append(result, getFileAttribute(c))
		}
	}
	return result
}

func getAttributeType(c *ishell.Context) msg.DocumentAttributeType {
	attribType := msg.AttributeTypeNone
	for {
		//AttributeTypeNone  DocumentAttributeType = 0
		//AttributeTypeAudio DocumentAttributeType = 1
		//AttributeTypeVideo DocumentAttributeType = 2
		//AttributeTypePhoto DocumentAttributeType = 3
		//AttributeTypeFile  DocumentAttributeType = 4
		//AttributeAnimated  DocumentAttributeType = 5
		c.Print("AttributeType : (Audio=1, Video=2 ,Photo=3, File=4 )")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil && id >= 0 && id < 5 {
			attribType = msg.DocumentAttributeType(id)
			break
		} else {
			c.Println("entered value is invalid ")
		}
	}
	return attribType
}

func getAudioAttribute(c *ishell.Context) *msg.DocumentAttribute {
	req := new(msg.DocumentAttribute)
	req.Type = msg.AttributeTypeAudio
	attrib := new(msg.DocumentAttributeAudio)
	attrib.Performer = fnGetPerformer(c)
	attrib.Title = fnGetTitle(c)
	attrib.Voice = fnGetVoice(c)
	attrib.Waveform = fnGetWaveForm(c)
	req.Data, _ = attrib.Marshal()
	return req
}
func getVideoAttribute(c *ishell.Context) *msg.DocumentAttribute {
	req := new(msg.DocumentAttribute)
	req.Type = msg.AttributeTypeVideo
	attrib := new(msg.DocumentAttributeVideo)
	attrib.Round = fnGetRound(c)
	attrib.Width = fnGetWidth(c)
	attrib.Height = fnGetHeight(c)
	req.Data, _ = attrib.Marshal()
	return req
}
func getPhotoAttribute(c *ishell.Context) *msg.DocumentAttribute {
	req := new(msg.DocumentAttribute)
	req.Type = msg.AttributeTypePhoto
	attrib := new(msg.DocumentAttributePhoto)
	attrib.Width = fnGetWidth(c)
	attrib.Height = fnGetHeight(c)
	req.Data, _ = attrib.Marshal()
	return req
}
func getFileAttribute(c *ishell.Context) *msg.DocumentAttribute {
	req := new(msg.DocumentAttribute)
	req.Type = msg.AttributeTypeFile
	attrib := new(msg.DocumentAttributeFile)
	attrib.Filename = fnGetFileName(c)
	req.Data, _ = attrib.Marshal()
	return req
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
	mediaType := msg.ClientMediaNone
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

func fnGetMediaTypes(c *ishell.Context) string {
	c.Print("Insert Media Types comma separated: Audio 1, Video 2, Photo 3, File 4, Animated 5")
	t := c.ReadLine()
	return t
}

func fnClearAll(c *ishell.Context) bool {
	del := false
	for {
		c.Print("clear all? : (0 = false , >=1 : true)")
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
