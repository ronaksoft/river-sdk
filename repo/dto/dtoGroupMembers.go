package dto

type GroupMembers struct {
	dto
	GroupID int64 `gorm:"type:bigint;primary_key;column:GroupID" json:"GroupID"` // type is required for composite primary key
	UserID  int32 `gorm:"type:bigint;primary_key;column:UserID" json:"UserID"`   // type is required for composite primary key
	Flags   int64 `gorm:"column:Flags" json:"Flags"`                             // reserved for creator|admin|restricted|blocked member and etc ...
}

func (GroupMembers) TableName() string {
	return "group_members"
}

// func (m *GroupMembers) Map(v *msg.XXXXXXXXX) {
// 	m.ID = v.ID
// 	m.CreatedOn = v.CreatedOn
// 	m.EditedOn = v.EditedOn
// 	m.Title = v.Title
// }

// func (m *GroupMembers) MapTo(v *msg.XXXXXXXXX) {
// 	v.ID = m.ID
// 	v.CreatedOn = m.CreatedOn
// 	v.EditedOn = m.EditedOn
// 	v.Title = m.Title
// }
