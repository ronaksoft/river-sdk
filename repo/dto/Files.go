package dto

import (
	"encoding/json"

	"git.ronaksoftware.com/ronak/riversdk/msg"
)

type Files struct {
	MessageID     int64  `gorm:"primary_key;column:MessageID;auto_increment:false" json:"MessageID"`
	PeerID        int64  `gorm:"column:PeerID" json:"PeerID"`
	PeerType      int32  `gorm:"column:PeerType" json:"PeerType"`
	ClusterID     int64  `gorm:"column:ClusterID" json:"ClusterID"`
	DocumentID    int64  `gorm:"column:DocumentID" json:"DocumentID"`
	AccessHash    int64  `gorm:"column:AccessHash" json:"AccessHash"`
	CreatedOn     int64  `gorm:"column:CreatedOn" json:"CreatedOn"`
	MediaType     int32  `gorm:"column:MediaType" json:"MediaType"`
	Caption       string `gorm:"type:TEXT;column:Caption" json:"Caption"`
	FileName      string `gorm:"type:TEXT;column:FileName" json:"FileName"`
	FilePath      string `gorm:"type:TEXT;column:FilePath" json:"FilePath"`
	ThumbFilePath string `gorm:"type:TEXT;column:ThumbFilePath" json:"ThumbFilePath"`
	FileMIME      string `gorm:"type:TEXT;column:FileMIME" json:"FileMIME"`
	ThumbMIME     string `gorm:"type:TEXT;column:ThumbMIME" json:"ThumbMIME"`
	ReplyTo       int64  `gorm:"column:ReplyTo" json:"ReplyTo"`
	ClearDraft    bool   `gorm:"column:ClearDraft" json:"ClearDraft"`
	Attributes    []byte `gorm:"type:blob;column:Attributes" json:"Attributes"`
}

func (Files) TableName() string {
	return "files"
}

func (m *Files) Map(messageID int64, createdOn int64, v *msg.ClientSendMessageMedia) {
	m.MessageID = messageID
	m.PeerID = v.Peer.ID
	m.PeerType = int32(v.Peer.Type)
	//m.ClusterID = v.ClusterID
	//m.DocumentID = v.DocumentID
	m.AccessHash = int64(v.Peer.AccessHash)
	m.MediaType = int32(v.MediaType)
	m.Caption = v.Caption
	m.FileName = v.FileName
	m.FilePath = v.FilePath
	m.ThumbFilePath = v.ThumbFilePath
	m.FileMIME = v.FileMIME
	m.ThumbMIME = v.ThumbMIME
	m.ReplyTo = v.ReplyTo
	m.ClearDraft = v.ClearDraft
	m.Attributes, _ = json.Marshal(v.Attributes)
}
