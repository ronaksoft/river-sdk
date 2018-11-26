package dto

import "git.ronaksoftware.com/ronak/riversdk/msg"

type GroupParticipants struct {
	dto
	GroupID   int64 `gorm:"type:bigint;primary_key;column:GroupID" json:"GroupID"` // type is required for composite primary key
	UserID    int64 `gorm:"type:bigint;primary_key;column:UserID" json:"UserID"`   // type is required for composite primary key
	InviterID int64 `gorm:"column:InviterID" json:"InviterID"`
	Date      int64 `gorm:"column:Date" json:"Date"`
	Type      int32 `gorm:"column:Type" json:"Type"`
}

func (GroupParticipants) TableName() string {
	return "group_participants"
}

func (m *GroupParticipants) MapFrom(groupID int64, v *msg.GroupParticipant) {
	m.GroupID = groupID
	m.UserID = v.UserID
	m.InviterID = v.InviterID
	m.Date = v.Date
	m.Type = int32(v.Type)
}

func (m *GroupParticipants) MapFromUpdateGroupMemberAdded(v *msg.UpdateGroupMemberAdded) {
	m.GroupID = v.ChatID
	m.UserID = v.UserID
	m.InviterID = v.InviterID
	m.Date = v.Date
	// m.Type = int32(v.Type)
	m.Type = int32(msg.ParticipantType_Member)
}

func (m *GroupParticipants) MapTo(v *msg.GroupParticipant) {
	v.UserID = m.UserID
	v.InviterID = m.InviterID
	v.Date = m.Date
	v.Type = msg.ParticipantType(m.Type)
}
