package dto

import (
	"fmt"
	"strconv"
	"strings"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
)

type Groups struct {
	dto
	ID           int64  `gorm:"primary_key;column:ID;auto_increment:false" json:"ID"`
	CreatedOn    int64  `gorm:"column:CreatedOn" json:"CreatedOn"`
	EditedOn     int64  `gorm:"column:EditedOn" json:"EditedOn"`
	Title        string `gorm:"type:TEXT;column:Title" json:"Title"`
	Participants int32  `gorm:"column:Participants" json:"Participants"`
	Flags        string `gorm:"column:Flags" json:"Flags"`

	Photo           []byte `gorm:"type:blob;column:Photo" json:"Photo"`
	BigFileID       int64  `gorm:";column:BigFileID" json:"BigFileID"`
	BigAccessHash   int64  `gorm:";column:BigAccessHash" json:"BigAccessHash"`
	BigClusterID    int32  `gorm:"column:BigClusterID" json:"BigClusterID"`
	BigVersion      int32  `gorm:"column:BigVersion" json:"BigVersion"`
	BigFilePath     string `gorm:"column:BigFilePath" json:"BigFilePath"`
	SmallFileID     int64  `gorm:";column:SmallFileID" json:"SmallFileID"`
	SmallAccessHash int64  `gorm:";column:SmallAccessHash" json:"SmallAccessHash"`
	SmallClusterID  int32  `gorm:"column:SmallClusterID" json:"SmallClusterID"`
	SmallVersion    int32  `gorm:"column:SmallVersion" json:"SmallVersion"`
	SmallFilePath   string `gorm:"column:SmallFilePath" json:"SmallFilePath"`
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
		m.SmallAccessHash = int64(v.Photo.PhotoSmall.AccessHash)
		m.SmallFileID = v.Photo.PhotoSmall.FileID
		m.SmallClusterID = v.Photo.PhotoSmall.ClusterID
		m.SmallVersion = 0

		m.BigAccessHash = int64(v.Photo.PhotoBig.AccessHash)
		m.BigFileID = v.Photo.PhotoBig.FileID
		m.BigClusterID = v.Photo.PhotoBig.ClusterID
		m.BigVersion = 0
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
		m.SmallAccessHash = int64(v.Photo.PhotoSmall.AccessHash)
		m.SmallFileID = v.Photo.PhotoSmall.FileID
		m.SmallClusterID = v.Photo.PhotoSmall.ClusterID
		m.SmallVersion = 0

		m.BigAccessHash = int64(v.Photo.PhotoBig.AccessHash)
		m.BigFileID = v.Photo.PhotoBig.FileID
		m.BigClusterID = v.Photo.PhotoBig.ClusterID
		m.BigVersion = 0
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
