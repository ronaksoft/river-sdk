package dto

type FileStatus struct {
	FileID     int64  `gorm:"primary_key;column:PeerID" json:"FileID"`
	FilePath   string `gorm:"column:FilePath" json:"FilePath"`
	Position   int64  `gorm:"column:Position" json:"Position"`
	TotalSize  int64  `gorm:"column:TotalSize" json:"TotalSize"`
	PartNo     int32  `gorm:"column:PartNo" json:"PartNo"`
	TotalParts int32  `gorm:"column:TotalParts" json:"TotalParts"`
	Completed  bool   `gorm:"column:Completed" json:"Completed"`
}

func (FileStatus) TableName() string {
	return "filestatus"
}
