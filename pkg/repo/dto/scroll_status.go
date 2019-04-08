package dto

type ScrollStatus struct {
	dto
	PeerID    int64 `gorm:"column:PeerID" json:"PeerID"`
	MessageID int64 `gorm:"column:MessageID" json:"MessageID"`
}

func (ScrollStatus) TableName() string {
	return "scrollstatus"
}
