package dto

type System struct {
	dto
	KeyName  string `gorm:"type:TEXT;primary_key;column:KeyName" json:"KeyName"`
	StrValue string `gorm:"type:TEXT;column:StrValue" json:"StrValue"`
	IntValue int32  `gorm:"column:IntValue" json:"IntValue"`
}

func (System) TableName() string {
	return "system"
}
