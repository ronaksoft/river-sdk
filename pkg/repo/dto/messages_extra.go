package dto

type MessagesExtra struct {
	dto
	PeerID    int64 `gorm:"column:PeerID" json:"PeerID"`
	PeerType  int32 `gorm:"column:PeerType" json:"PeerType"`
	ScrollID int64 `gorm:"column:ScrollID" json:"ScrollID"`
}

func (MessagesExtra) TableName() string {
	return "messages_extra"
}
