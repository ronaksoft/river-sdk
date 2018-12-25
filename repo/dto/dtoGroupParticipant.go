package dto

import "git.ronaksoftware.com/ronak/riversdk/msg"

type GroupParticipants struct {
	dto
	GroupID    int64  `gorm:"type:bigint;primary_key;column:GroupID" json:"GroupID"` // type is required for composite primary key
	UserID     int64  `gorm:"type:bigint;primary_key;column:UserID" json:"UserID"`   // type is required for composite primary key
	FirstName  string `gorm:"column:FirstName" json:"FirstName"`
	LastName   string `gorm:"column:LastName" json:"LastName"`
	Type       int32  `gorm:"column:Type" json:"Type"`
	AccessHash int64  `gorm:"column:AccessHash" json:"AccessHash"`
	Username   string `gorm:"column:Username" json:"Username"`
}

func (GroupParticipants) TableName() string {
	return "group_participants"
}

func (m *GroupParticipants) MapFrom(groupID int64, v *msg.GroupParticipant) {
	m.GroupID = groupID
	m.UserID = v.UserID
	m.FirstName = v.FirstName
	m.LastName = v.LastName
	m.Type = int32(v.Type)
	m.AccessHash = int64(v.AccessHash)
	m.Username = v.Username
}

func (m *GroupParticipants) MapTo(v *msg.GroupParticipant) {
	v.UserID = m.UserID
	v.FirstName = m.FirstName
	v.LastName = m.LastName
	v.Type = msg.ParticipantType(m.Type)
	v.AccessHash = uint64(m.AccessHash)
	v.Username = v.Username
}
