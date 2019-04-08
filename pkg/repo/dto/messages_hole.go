package dto

import (
	"database/sql"
)

type MessagesHole struct {
	dto
	PeerID int64         `gorm:"type:bigint;primary_key;column:PeerID" json:"PeerID"`
	MinID  sql.NullInt64 `gorm:"default:0;type:bigint;primary_key;column:MinID" json:"MinID"`
	MaxID  int64         `gorm:"type:bigint;column:MaxID" json:"MaxID"`
}

func (MessagesHole) TableName() string {
	return "messages_hole"
}
