package dto

import (
	"fmt"
	"strconv"
	"strings"

	"git.ronaksoftware.com/ronak/riversdk/msg"
)

type Groups struct {
	dto
	ID           int64  `gorm:"primary_key;column:ID;auto_increment:false" json:"ID"`
	CreatedOn    int64  `gorm:"column:CreatedOn" json:"CreatedOn"`
	EditedOn     int64  `gorm:"column:EditedOn" json:"EditedOn"`
	Title        string `gorm:"type:TEXT;column:Title" json:"Title"`
	Participants int32  `gorm:"column:Participants" json:"Participants"`
	Flags        string `gorm:"column:Flags" json:"Flags"`
}

func (Groups) TableName() string {
	return "groups"
}

func (m *Groups) MapFrom(v *msg.Group) {
	m.ID = v.ID
	m.CreatedOn = v.CreatedOn
	m.EditedOn = v.EditedOn
	m.Title = v.Title
	m.Participants = v.Participants
	m.Flags = fnFlagsToString(v.Flags)
}

func (m *Groups) MapTo(v *msg.Group) {
	v.ID = m.ID
	v.CreatedOn = m.CreatedOn
	v.EditedOn = m.EditedOn
	v.Title = m.Title
	v.Participants = m.Participants
	v.Flags = fnFlagsToArray(m.Flags)
}

func fnFlagsToString(flags []msg.GroupFlags) string {
	sb := new(strings.Builder)
	for _, f := range flags {
		sb.WriteString(fmt.Sprintf("%d;", int32(f)))
	}
	return sb.String()
}
func fnFlagsToArray(flags string) []msg.GroupFlags {
	res := make([]msg.GroupFlags, 0)
	strFlags := strings.Split(flags, ";")
	for _, s := range strFlags {
		trimS := strings.TrimSpace(s)
		if trimS == "" {
			continue
		}
		tmp, err := strconv.Atoi(trimS)
		if err == nil {
			res = append(res, msg.GroupFlags(tmp))
		}
	}

	return res
}
