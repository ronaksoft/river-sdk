package dto

type MessagesExtra struct {
	dto
	PeerID   int64 `gorm:"type:bigint;primary_key;column:PeerID" json:"PeerID"`
	PeerType int32 `gorm:"type:integer;primary_key;column:PeerType" json:"PeerType"`
	ScrollID int64 `gorm:"column:ScrollID" json:"ScrollID"`
}

func (MessagesExtra) TableName() string {
	return "messages_extra"
}
