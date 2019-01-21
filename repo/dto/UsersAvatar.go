package dto

type UsersAvatar struct {
	dto
	UserID     int64 `gorm:"type:bigint;primary_key;column:UserID" json:"UserID"`  //type is required for composite primary key
	FileID     int64 `gorm:"type:integer;primary_key;column:FileID" json:"FileID"` //type is required for composite primary key
	ClusterID  int64 `gorm:"column:ClusterID" json:"ClusterID"`
	Version    int32 `gorm:"column:Version" json:"Version"`
	FilePath   int64 `gorm:"column:FilePath" json:"FilePath"`
	PhotoSmall bool  `gorm:"column:PhotoSmall" json:"PhotoSmall"`
	PhotoBig   bool  `gorm:"column:PhotoBig" json:"PhotoBig"`
}

func (UsersAvatar) TableName() string {
	return "users_avatar"
}

func (m *UsersAvatar) Map() {

}
func (m *UsersAvatar) MapTo() {

}
