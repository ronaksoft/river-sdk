package dto

import "git.ronaksoftware.com/ronak/riversdk/msg"

type UserPhotos struct {
	dto
	UserID           int64  `gorm:"type:bigint;primary_key;column:UserID" json:"UserID"`   //type is required for composite primary key
	PhotoID          int64  `gorm:"type:bigint;primary_key;column:PhotoID" json:"PhotoID"` //type is required for composite primary key
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

func (UserPhotos) TableName() string {
	return "user_photos"
}

func (m *UserPhotos) Map(userId int64, v *msg.UserPhoto) {
	m.UserID = userId
	m.PhotoID = v.PhotoID
	m.Big_FileID = v.PhotoBig.FileID
	m.Big_AccessHash = int64(v.PhotoBig.AccessHash)
	m.Big_ClusterID = v.PhotoBig.ClusterID
	//m.Big_Version = v.PhotoBig.Version
	m.Big_FilePath = ""
	m.Small_FileID = v.PhotoSmall.FileID
	m.Small_AccessHash = int64(v.PhotoSmall.AccessHash)
	m.Small_ClusterID = v.PhotoSmall.ClusterID
	//m.Small_Version = v.PhotoSmall.Version
	m.Small_FilePath = ""

}
func (m *UserPhotos) MapTo() {

}
