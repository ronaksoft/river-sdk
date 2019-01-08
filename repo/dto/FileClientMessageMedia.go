package dto

import (
	"encoding/json"

	"git.ronaksoftware.com/ronak/riversdk/msg"
)

type FileClientMessageMedia struct {
	FileID        int64  `gorm:"primary_key;column:FileID;auto_increment:false" json:"FileID"`
	PeerID        int64  `gorm:"column:PeerID" json:"PeerID"`
	PeerType      int32  `gorm:"column:PeerType" json:"PeerType"`
	AccessHash    int64  `gorm:"column:AccessHash" json:"AccessHash"`
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

func (FileClientMessageMedia) TableName() string {
	return "fileclientmessagemedia"
}

func (m *FileClientMessageMedia) Map(fileID int64, v *msg.ClientSendMessageMedia) {
	m.FileID = fileID
	m.PeerID = v.Peer.ID
	m.PeerType = int32(v.Peer.Type)
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

func (m *FileClientMessageMedia) MapTo(v *msg.ClientSendMessageMedia) {
	if v.Peer == nil {
		v.Peer = new(msg.InputPeer)
	}
	v.Peer.ID = m.PeerID
	v.Peer.Type = msg.PeerType(m.PeerType)
	v.Peer.AccessHash = uint64(m.AccessHash)
	v.MediaType = msg.InputMediaType(m.MediaType)
	v.Caption = m.Caption
	v.FileName = m.FileName
	v.FilePath = m.FilePath
	v.ThumbFilePath = m.ThumbFilePath
	v.FileMIME = m.FileMIME
	v.ThumbMIME = m.ThumbMIME
	v.ReplyTo = m.ReplyTo
	v.ClearDraft = m.ClearDraft
	if v.Attributes == nil {
		v.Attributes = make([]*msg.DocumentAttribute, 0)
	}
	json.Unmarshal(m.Attributes, &v.Attributes)
}
