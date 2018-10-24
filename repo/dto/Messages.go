package dto

import "git.ronaksoftware.com/ronak/riversdk/msg"

type Messages struct {
	dto
	ID                  int64  `gorm:"primary_key;column:ID;auto_increment:false" json:"ID"`
	PeerID              int64  `gorm:"column:PeerID" json:"PeerID"`
	PeerType            int32  `gorm:"column:PeerType" json:"PeerType"`
	CreatedOn           int64  `gorm:"column:CreatedOn" json:"CreatedOn"`
	EditedOn            int64  `gorm:"column:EditedOn" json:"EditedOn"`
	FwdSenderID         int64  `gorm:"column:FwdSenderID" json:"FwdSenderID"`
	FwdChannelID        int64  `gorm:"column:FwdChannelID" json:"FwdChannelID"`
	FwdChannelMessageID int64  `gorm:"column:FwdChannelMessageID" json:"FwdChannelMessageID"`
	Flags               int32  `gorm:"column:Flags" json:"Flags"`
	MessageType         int64  `gorm:"column:MessageType" json:"MessageType"`
	Body                string `gorm:"type:TEXT;column:Body" json:"Body"`
	SenderID            int64  `gorm:"column:SenderID" json:"SenderID"`
	ContentRead         bool   `gorm:"column:ContentRead" json:"ContentRead"`
	Inbox               bool   `gorm:"column:Inbox" json:"Inbox"`
	ReplyTo             int64  `gorm:"column:ReplyTo" json:"ReplyTo"`
	MessageAction       int32  `gorm:"column:MessageAction" json:"MessageAction"`
}

func (Messages) TableName() string {
	return "messages"
}

func (m *Messages) Map(v *msg.UserMessage) {
	m.ID = v.ID
	m.PeerID = v.PeerID
	m.PeerType = v.PeerType
	m.CreatedOn = v.CreatedOn
	m.EditedOn = v.EditedOn
	m.FwdSenderID = v.FwdSenderID
	m.FwdChannelID = v.FwdChannelID
	m.FwdChannelMessageID = v.FwdChannelMessageID
	m.Flags = v.Flags
	m.MessageType = v.MessageType
	m.Body = v.Body
	m.SenderID = v.SenderID
	m.ContentRead = v.ContentRead
	m.Inbox = v.Inbox
	m.ReplyTo = v.ReplyTo
	m.MessageAction = v.MessageAction
}

func (m *Messages) MapTo(v *msg.UserMessage) {
	v.ID = m.ID
	v.PeerID = m.PeerID
	v.PeerType = m.PeerType
	v.CreatedOn = m.CreatedOn
	v.EditedOn = m.EditedOn
	v.FwdSenderID = m.FwdSenderID
	v.FwdChannelID = m.FwdChannelID
	v.FwdChannelMessageID = m.FwdChannelMessageID
	v.Flags = m.Flags
	v.MessageType = m.MessageType
	v.Body = m.Body
	v.SenderID = m.SenderID
	v.ContentRead = m.ContentRead
	v.Inbox = m.Inbox
	v.ReplyTo = m.ReplyTo
	v.MessageAction = m.MessageAction
}
