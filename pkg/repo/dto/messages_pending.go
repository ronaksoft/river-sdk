package dto

import (
	"encoding/json"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
)

const (
	_ClientSendMessageMediaType       = -1
	_ClientSendMessageContactType     = -2
	_ClientSendMessageGeoLocationType = -3
)

type MessagesPending struct {
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

func (MessagesPending) TableName() string {
	return "messages_pending"
}

func (m *MessagesPending) Map(v *msg.MessagesSend) {
	m.AccessHash = int64(v.Peer.AccessHash)
	m.Body = v.Body
	m.PeerID = v.Peer.ID
	m.PeerType = int32(v.Peer.Type)
	m.ReplyTo = v.ReplyTo
	m.RequestID = v.RandomID
	m.ClearDraft = v.ClearDraft
	m.Entities, _ = json.Marshal(v.Entities)
}

func (m *MessagesPending) MapTo(v *msg.ClientPendingMessage) {
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
func (m *MessagesPending) MapToUserMessage(v *msg.UserMessage) {
	v.ID = m.ID
	v.PeerID = m.PeerID
	v.PeerType = m.PeerType
	v.CreatedOn = m.CreatedOn
	v.Body = m.Body
	v.SenderID = m.SenderID
	v.ReplyTo = m.ReplyTo

	v.Entities = make([]*msg.MessageEntity, 0)
	json.Unmarshal(m.Entities, &v.Entities)

	v.MessageType = fnGetMessageType(m.MediaType)

	v.MediaType = msg.MediaType(m.MediaType)
	v.Media = m.Media
}

func (m *MessagesPending) MapToDtoMessage(v *Messages) {
	v.ID = m.ID
	v.PeerID = m.PeerID
	v.PeerType = m.PeerType
	v.CreatedOn = m.CreatedOn
	v.Body = m.Body
	v.SenderID = m.SenderID
	v.ReplyTo = m.ReplyTo
	v.Entities = m.Entities
	v.MessageType = fnGetMessageType(m.MediaType)
	v.MediaType = int32(m.MediaType)
	v.Media = m.Media

}

func (m *MessagesPending) MapToMessageSend(v *msg.MessagesSend) {
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

func (m *MessagesPending) MapFromClientMessageMedia(fileID int64, v *msg.ClientSendMessageMedia) {
	m.PeerID = v.Peer.ID
	m.PeerType = int32(v.Peer.Type)
	m.AccessHash = int64(v.Peer.AccessHash)
	m.Body = v.Caption
	m.ReplyTo = v.ReplyTo
	m.ClearDraft = v.ClearDraft
	m.MediaType = int32(v.MediaType)
	m.Media, _ = v.Marshal()
}
func (m *MessagesPending) MapFromMessageMedia(v *msg.MessagesSendMedia) {
	m.RequestID = v.RandomID
	m.PeerID = v.Peer.ID
	m.PeerType = int32(v.Peer.Type)
	m.AccessHash = int64(v.Peer.AccessHash)
	m.ReplyTo = v.ReplyTo
	m.ClearDraft = v.ClearDraft
	m.MediaType = int32(v.MediaType)
	m.Media = v.MediaData
}

func fnGetMessageType(inputType int32) int64 {
	if inputType == int32(msg.InputMediaTypeUploadedDocument) {
		return _ClientSendMessageMediaType
	}
	if inputType == int32(msg.InputMediaTypeContact) {
		return _ClientSendMessageContactType
	}
	if inputType == int32(msg.InputMediaTypeGeoLocation) {
		return _ClientSendMessageGeoLocationType
	}
	return 0
}
