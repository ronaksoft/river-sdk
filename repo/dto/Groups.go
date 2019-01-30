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

	Photo            []byte `gorm:"type:blob;column:Photo" json:"Photo"`
	Big_FileID       int64  `gorm:";column:Big_FileID" json:"Big_FileID"`
	Big_AccessHash   int64  `gorm:";column:Big_AccessHash" json:"Big_AccessHash"`
	Big_ClusterID    int32  `gorm:"column:Big_ClusterID" json:"Big_ClusterID"`
	Big_Version      int32  `gorm:"column:Big_Version" json:"Big_Version"`
	Big_FilePath     string `gorm:"column:Big_FilePath" json:"Big_FilePath"`
	Small_FileID     int64  `gorm:";column:Small_FileID" json:"Small_FileID"`
	Small_AccessHash int64  `gorm:";column:Small_AccessHash" json:"Small_AccessHash"`
	Small_ClusterID  int32  `gorm:"column:Small_ClusterID" json:"Small_ClusterID"`
	Small_Version    int32  `gorm:"column:Small_Version" json:"Small_Version"`
	Small_FilePath   string `gorm:"column:Small_FilePath" json:"Small_FilePath"`
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
	if v.Photo != nil {
		m.Photo, _ = v.Photo.Marshal()
		m.Small_AccessHash = int64(v.Photo.PhotoSmall.AccessHash)
		m.Small_AccessHash = v.Photo.PhotoSmall.FileID
		m.Small_ClusterID = v.Photo.PhotoSmall.ClusterID
		m.Small_Version = 0

		m.Big_AccessHash = int64(v.Photo.PhotoBig.AccessHash)
		m.Big_AccessHash = v.Photo.PhotoBig.FileID
		m.Big_ClusterID = v.Photo.PhotoBig.ClusterID
		m.Big_Version = 0
	}
}

func (m *Groups) MapTo(v *msg.Group) {
	v.ID = m.ID
	v.CreatedOn = m.CreatedOn
	v.EditedOn = m.EditedOn
	v.Title = m.Title
	v.Participants = m.Participants
	v.Flags = fnFlagsToArray(m.Flags)
	if v.Photo == nil {
		v.Photo = new(msg.GroupPhoto)
	}
	err := v.Photo.Unmarshal(m.Photo)
	if err != nil {
		v.Photo = nil
	}
}

func (m *Groups) MapFromUpdateGroupPhoto(v *msg.UpdateGroupPhoto) {
	m.ID = v.GroupID
	if v.Photo != nil {
		m.Photo, _ = v.Photo.Marshal()
		m.Small_AccessHash = int64(v.Photo.PhotoSmall.AccessHash)
		m.Small_AccessHash = v.Photo.PhotoSmall.FileID
		m.Small_ClusterID = v.Photo.PhotoSmall.ClusterID
		m.Small_Version = 0

		m.Big_AccessHash = int64(v.Photo.PhotoBig.AccessHash)
		m.Big_AccessHash = v.Photo.PhotoBig.FileID
		m.Big_ClusterID = v.Photo.PhotoBig.ClusterID
		m.Big_Version = 0
	}
}

func fnFlagsToString(flags []msg.GroupFlags) string {
	sb := new(strings.Builder)
	for _, f := range flags {
		sb.WriteString(fmt.Sprintf("%d;", int32(f)))
	}
	sb.WriteString(";") // prevent leaving empty flags cuz it wont be saved
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
