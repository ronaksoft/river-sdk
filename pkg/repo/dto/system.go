package dto

type System struct {
	dto
	KeyName  string `gorm:"type:TEXT;primary_key;column:KeyName" json:"KeyName"`
	StrValue string `gorm:"type:TEXT;column:StrValue" json:"StrValue"`
	IntValue int32  `gorm:"column:IntValue" json:"IntValue"`
	Salt     string `gorm:"column:Salt" json:"Salt"`
}

func (System) TableName() string {
	return "system"
}

type ServerSalt struct {
	Timestamp int64 `json:"timestamp`
	Salt      int64 `json:"salt`
}
