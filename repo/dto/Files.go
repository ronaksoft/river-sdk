package dto

import (
	"encoding/json"

	"git.ronaksoftware.com/ronak/riversdk/msg"
)

type Files struct {
	MessageID     int64  `gorm:"primary_key;column:MessageID;auto_increment:false" json:"MessageID"`
	PeerID        int64  `gorm:"column:PeerID" json:"PeerID"`
	PeerType      int32  `gorm:"column:PeerType" json:"PeerType"`
	ClusterID     int32  `gorm:"column:ClusterID" json:"ClusterID"`
	DocumentID    int64  `gorm:"column:DocumentID" json:"DocumentID"`
	AccessHash    int64  `gorm:"column:AccessHash" json:"AccessHash"`
	CreatedOn     int64  `gorm:"column:CreatedOn" json:"CreatedOn"`
	MediaType     int32  `gorm:"column:MediaType" json:"MediaType"`
	Caption       string `gorm:"type:TEXT;column:Caption" json:"Caption"`
	FileName      string `gorm:"type:TEXT;column:FileName" json:"FileName"`
	FileSize      int32  `gorm:"column:FileSize" json:"FileSize"`
	FilePath      string `gorm:"type:TEXT;column:FilePath" json:"FilePath"`
	ThumbFilePath string `gorm:"type:TEXT;column:ThumbFilePath" json:"ThumbFilePath"`
	FileMIME      string `gorm:"type:TEXT;column:FileMIME" json:"FileMIME"`
	ThumbMIME     string `gorm:"type:TEXT;column:ThumbMIME" json:"ThumbMIME"`
	ReplyTo       int64  `gorm:"column:ReplyTo" json:"ReplyTo"`
	ClearDraft    bool   `gorm:"column:ClearDraft" json:"ClearDraft"`
	Version       int32  `gorm:"column:Version" json:"Version"`
	Attributes    []byte `gorm:"type:blob;column:Attributes" json:"Attributes"`
}

func (Files) TableName() string {
	return "files"
}

func (m *Files) Map(messageID int64, createdOn int64, fileSize int32, v *msg.ClientSendMessageMedia) {
	m.MessageID = messageID
	m.PeerID = v.Peer.ID
	m.PeerType = int32(v.Peer.Type)
	//m.ClusterID = v.ClusterID
	//m.DocumentID = v.DocumentID
	m.AccessHash = int64(v.Peer.AccessHash)
	m.MediaType = int32(v.MediaType)
	m.Caption = v.Caption
	m.FileName = v.FileName
	m.FileSize = fileSize
	m.FilePath = v.FilePath
	m.ThumbFilePath = v.ThumbFilePath
	m.FileMIME = v.FileMIME
	m.ThumbMIME = v.ThumbMIME
	m.ReplyTo = v.ReplyTo
	m.ClearDraft = v.ClearDraft
	m.Version = 0
	m.Attributes, _ = json.Marshal(v.Attributes)
}

func (m *Files) MapFromFile(v Files) {
	m.MessageID = v.MessageID
	m.PeerID = v.PeerID
	m.PeerType = v.PeerType
	m.ClusterID = v.ClusterID
	m.DocumentID = v.DocumentID
	m.AccessHash = v.AccessHash
	m.CreatedOn = v.CreatedOn
	m.MediaType = v.MediaType
	m.Caption = v.Caption
	m.FileName = v.FileName
	m.FileSize = v.FileSize
	m.FilePath = v.FilePath
	m.ThumbFilePath = v.ThumbFilePath
	m.FileMIME = v.FileMIME
	m.ThumbMIME = v.ThumbMIME
	m.ReplyTo = v.ReplyTo
	m.ClearDraft = v.ClearDraft
	m.Version = v.Version
	m.Attributes = v.Attributes
}

func (m *Files) MapFromDocument(v *msg.MediaDocument) {
	//m.MessageID = msgID
	m.Caption = v.Caption
	m.DocumentID = v.Doc.ID
	m.AccessHash = int64(v.Doc.AccessHash)
	m.CreatedOn = v.Doc.Date
	m.FileMIME = v.Doc.MimeType
	m.FileSize = v.Doc.FileSize
	m.Version = v.Doc.Version
	m.ClusterID = v.Doc.ClusterID
	m.Attributes, _ = json.Marshal(v.Doc.Attributes)
}

func (m *Files) MapFromFileStatus(v *FileStatus) {
	m.MessageID = v.MessageID
	m.DocumentID = v.FileID
	m.ClusterID = v.ClusterID
	m.AccessHash = v.AccessHash
	m.Version = v.Version
	m.FilePath = v.FilePath
	//m.Position = v.Position
	m.FileSize = int32(v.TotalSize)
	//m.PartNo = v.PartNo
	//m.TotalParts = v.TotalParts
	//m.IsCompleted = v.IsCompleted
	if v.Type {
		// Download state
		doc := new(msg.Document)
		err := doc.Unmarshal(v.DownloadRequest)
		if err == nil {
			m.DocumentID = doc.ID
			m.AccessHash = int64(doc.AccessHash)
			m.CreatedOn = doc.Date
			m.FileMIME = doc.MimeType
			m.FileSize = doc.FileSize
			m.Version = doc.Version
			m.ClusterID = doc.ClusterID
			m.Attributes, _ = json.Marshal(doc.Attributes)
		}

	} else {
		// upload state
		req := new(msg.ClientSendMessageMedia)
		err := req.Unmarshal(v.UploadRequest)
		if err == nil {
			//m.MessageID = messageID
			m.PeerID = req.Peer.ID
			m.PeerType = int32(req.Peer.Type)
			//m.ClusterID = v.ClusterID
			//m.DocumentID = v.DocumentID
			m.AccessHash = int64(req.Peer.AccessHash)
			m.MediaType = int32(req.MediaType)
			m.Caption = req.Caption
			m.FileName = req.FileName
			//m.FileSize = fileSize
			m.FilePath = req.FilePath
			m.ThumbFilePath = req.ThumbFilePath
			m.FileMIME = req.FileMIME
			m.ThumbMIME = req.ThumbMIME
			m.ReplyTo = req.ReplyTo
			m.ClearDraft = req.ClearDraft
			//m.Version = 0
			m.Attributes, _ = json.Marshal(req.Attributes)
		}
	}
}
