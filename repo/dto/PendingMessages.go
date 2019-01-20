package dto

import (
	"encoding/json"

	"git.ronaksoftware.com/ronak/riversdk/msg"
)

const (
	_ClientSendMessageMediaType   = -1
	_ClientSendMessageContactType = -2
)

type PendingMessages struct {
	dto
	ID         int64  `gorm:"primary_key;column:ID;auto_increment:false" json:"ID"`
	RequestID  int64  `gorm:"column:RequestID;index:ixRequestID" json:"RequestID"`
	PeerID     int64  `gorm:"column:PeerID" json:"PeerID"`
	SenderID   int64  `gorm:"column:SenderID" json:"SenderID"`
	PeerType   int32  `gorm:"column:PeerType" json:"PeerType"`
	AccessHash int64  `gorm:"column:AccessHash" json:"AccessHash"`
	CreatedOn  int64  `gorm:"column:CreatedOn" json:"CreatedOn"`
	Body       string `gorm:"type:TEXT;column:Body" json:"Body"`
	ReplyTo    int64  `gorm:"column:ReplyTo" json:"ReplyTo"`
	ClearDraft bool   `gorm:"column:ClearDraft" json:"ClearDraft"`
	Entities   []byte `gorm:"type:blob;column:Entities" json:"Entities"`
	MediaType  int32  `gorm:"column:MediaType" json:"MediaType"`
	Media      []byte `gorm:"type:blob;column:Media" json:"Media"`
}

func (PendingMessages) TableName() string {
	return "pendingmessages"
}

func (m *PendingMessages) Map(v *msg.MessagesSend) {
	m.AccessHash = int64(v.Peer.AccessHash)
	m.Body = v.Body
	//m.CreatedOn = v.CreatedOn
	//m.ID = v.ID
	m.PeerID = v.Peer.ID
	m.PeerType = int32(v.Peer.Type)
	m.ReplyTo = v.ReplyTo
	m.RequestID = v.RandomID
	m.ClearDraft = v.ClearDraft
	m.Entities, _ = json.Marshal(v.Entities)
}

func (m *PendingMessages) MapTo(v *msg.ClientPendingMessage) {
	v.ID = m.ID
	v.RequestID = m.RequestID
	v.PeerID = m.PeerID
	v.PeerType = m.PeerType
	v.AccessHash = uint64(m.AccessHash)
	v.CreatedOn = m.CreatedOn
	v.ReplyTo = m.ReplyTo
	v.Body = m.Body
	v.SenderID = m.SenderID

	v.Entities = make([]*msg.MessageEntity, 0)
	json.Unmarshal(m.Entities, &v.Entities)

	v.MediaType = msg.InputMediaType(m.MediaType)
	v.Media = m.Media
}
func (m *PendingMessages) MapToUserMessage(v *msg.UserMessage) {
	v.ID = m.ID
	v.PeerID = m.PeerID
	v.PeerType = m.PeerType
	v.CreatedOn = m.CreatedOn
	//v.EditedOn = m.EditedOn
	//v.FwdSenderID = m.FwdSenderID
	//v.FwdChannelID = m.FwdChannelID
	//v.FwdChannelMessageID = m.FwdChannelMessageID
	//v.Flags = m.Flags
	//v.MessageType = m.MessageType
	v.Body = m.Body
	v.SenderID = m.SenderID
	//v.ContentRead = m.ContentRead
	//v.Inbox = m.Inbox
	v.ReplyTo = m.ReplyTo
	//v.MessageAction = m.MessageAction

	v.Entities = make([]*msg.MessageEntity, 0)
	json.Unmarshal(m.Entities, &v.Entities)

	if m.MediaType == int32(msg.InputMediaTypeUploadedDocument) {
		v.MessageType = _ClientSendMessageMediaType
	}
	if m.MediaType == int32(msg.InputMediaTypeContact) {
		v.MessageType = _ClientSendMessageContactType
	}

	v.MediaType = msg.MediaType(m.MediaType)
	v.Media = m.Media
}

func (m *PendingMessages) MapToDtoMessage(v *Messages) {
	v.ID = m.ID
	v.PeerID = m.PeerID
	v.PeerType = m.PeerType
	v.CreatedOn = m.CreatedOn
	//v.EditedOn = m.EditedOn
	//v.FwdSenderID = m.FwdSenderID
	//v.FwdChannelID = m.FwdChannelID
	//v.FwdChannelMessageID = m.FwdChannelMessageID
	//v.Flags = m.Flags
	//v.MessageType = m.MessageType
	v.Body = m.Body
	v.SenderID = m.SenderID
	//v.ContentRead = m.ContentRead
	//v.Inbox = m.Inbox
	v.ReplyTo = m.ReplyTo
	//v.MessageAction = m.MessageAction
	v.Entities = m.Entities

	if m.MediaType == int32(msg.InputMediaTypeUploadedDocument) {
		v.MessageType = _ClientSendMessageMediaType
	}
	if m.MediaType == int32(msg.InputMediaTypeContact) {
		v.MessageType = _ClientSendMessageContactType
	}
	v.MediaType = int32(m.MediaType)
	v.Media = m.Media

}

func (m *PendingMessages) MapToMessageSend(v *msg.MessagesSend) {

	v.Body = m.Body
	v.Peer = new(msg.InputPeer)
	v.Peer.AccessHash = uint64(m.AccessHash)
	v.Peer.ID = m.PeerID
	v.Peer.Type = msg.PeerType(m.PeerType)
	v.RandomID = m.RequestID
	v.ReplyTo = m.ReplyTo
	v.ClearDraft = m.ClearDraft

	v.Entities = make([]*msg.MessageEntity, 0)
	json.Unmarshal(m.Entities, &v.Entities)
}

func (m *PendingMessages) MapFromClientMessageMedia(fileID int64, v *msg.ClientSendMessageMedia) {
	//m.ID = v.ID
	//m.RequestID = fileID
	m.PeerID = v.Peer.ID
	//m.SenderID = v.
	m.PeerType = int32(v.Peer.Type)
	m.AccessHash = int64(v.Peer.AccessHash)
	//m.CreatedOn = v.
	m.Body = v.Caption
	m.ReplyTo = v.ReplyTo
	m.ClearDraft = v.ClearDraft
	// m.Entities = v.
	m.MediaType = int32(v.MediaType)
	m.Media, _ = v.Marshal()
}
func (m *PendingMessages) MapFromMessageMedia(v *msg.MessagesSendMedia) {
	m.RequestID = v.RandomID
	m.PeerID = v.Peer.ID
	m.PeerType = int32(v.Peer.Type)
	m.AccessHash = int64(v.Peer.AccessHash)
	m.ReplyTo = v.ReplyTo
	m.ClearDraft = v.ClearDraft
	m.MediaType = int32(v.MediaType)
	m.Media, _ = v.Marshal()
}
