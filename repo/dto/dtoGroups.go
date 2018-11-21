package dto

import (
	"git.ronaksoftware.com/ronak/riversdk/msg"
)

type Groups struct {
	dto
	ID        int64  `gorm:"primary_key;column:ID;auto_increment:false" json:"ID"`
	CreatedOn int64  `gorm:"column:CreatedOn" json:"CreatedOn"`
	EditedOn  int64  `gorm:"column:EditedOn" json:"EditedOn"`
	Title     string `gorm:"type:TEXT;column:Title" json:"Title"`
}

func (Groups) TableName() string {
	return "groups"
}

func (m *Groups) MapFrom(v *msg.Group) {
	m.ID = v.ID
	//m.CreatedOn = time.Now().Unix()
	//m.EditedOn = v.EditedOn
	m.Title = v.Title
}

func (m *Groups) MapTo(v *msg.Group) {
	v.ID = m.ID
	//v.CreatedOn = m.CreatedOn
	//v.EditedOn = m.EditedOn
	v.Title = m.Title
}
