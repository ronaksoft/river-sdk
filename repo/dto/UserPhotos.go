package dto

type UserPhotos struct {
	dto
	UserID     int64 `gorm:"type:bigint;primary_key;column:UserID" json:"UserID"`  //type is required for composite primary key
	FileID     int64 `gorm:"type:integer;primary_key;column:FileID" json:"FileID"` //type is required for composite primary key
	ClusterID  int64 `gorm:"column:ClusterID" json:"ClusterID"`
	Version    int32 `gorm:"column:Version" json:"Version"`
	FilePath   int64 `gorm:"column:FilePath" json:"FilePath"`
	PhotoSmall bool  `gorm:"column:PhotoSmall" json:"PhotoSmall"`
	PhotoBig   bool  `gorm:"column:PhotoBig" json:"PhotoBig"`
}

func (UserPhotos) TableName() string {
	return "user_photos"
}

func (m *UserPhotos) Map() {

}
func (m *UserPhotos) MapTo() {

}
