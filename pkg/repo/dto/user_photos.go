package dto

import "git.ronaksoftware.com/ronak/riversdk/msg"

type UserPhotos struct {
	dto
	UserID           int64  `gorm:"type:bigint;primary_key;column:UserID" json:"UserID"`   // type is required for composite primary key
	PhotoID         int64  `gorm:"type:bigint;primary_key;column:PhotoID" json:"PhotoID"` // type is required for composite primary key
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

func (UserPhotos) TableName() string {
	return "user_photos"
}

func (m *UserPhotos) Map(userId int64, v *msg.UserPhoto) {
	m.UserID = userId
	m.PhotoID = v.PhotoID
	m.BigFileID = v.PhotoBig.FileID
	m.BigAccessHash = int64(v.PhotoBig.AccessHash)
	m.BigClusterID = v.PhotoBig.ClusterID
	m.SmallFileID = v.PhotoSmall.FileID
	m.SmallAccessHash = int64(v.PhotoSmall.AccessHash)
	m.SmallClusterID = v.PhotoSmall.ClusterID
}

func (m *UserPhotos) MapTo(v *msg.UserPhoto) {
	v.PhotoID = m.PhotoID
	v.PhotoBig = new(msg.FileLocation)
	v.PhotoSmall = new(msg.FileLocation)
	v.PhotoBig.FileID = m.BigFileID
	v.PhotoBig.AccessHash = uint64(m.BigAccessHash)
	v.PhotoBig.ClusterID = m.BigClusterID
	v.PhotoSmall.FileID = m.SmallFileID
	v.PhotoSmall.AccessHash = uint64(m.SmallAccessHash)
	v.PhotoSmall.ClusterID = m.SmallClusterID
}
