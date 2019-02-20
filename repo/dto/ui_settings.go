package dto

type UISettings struct {
	dto
	Key   string `gorm:"type:TEXT;primary_key;column:Key" json:"Key"`
	Value string `gorm:"type:TEXT;column:Value" json:"Value"`
}

func (UISettings) TableName() string {
	return "uisettings"
}
