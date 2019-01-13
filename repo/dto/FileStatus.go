package dto

type FileStatus struct {
	MessageID       int64  `gorm:"primary_key;column:MessageID" json:"MessageID"`
	FileID          int64  `gorm:"column:FileID" json:"FileID"`
	ClusterID       int32  `gorm:"column:ClusterID" json:"ClusterID"`
	AccessHash      int64  `gorm:"column:AccessHash" json:"AccessHash"`
	Version         int32  `gorm:"column:Version" json:"Version"`
	FilePath        string `gorm:"column:FilePath" json:"FilePath"`
	Position        int64  `gorm:"column:Position" json:"Position"`
	TotalSize       int64  `gorm:"column:TotalSize" json:"TotalSize"`
	PartNo          int32  `gorm:"column:PartNo" json:"PartNo"`
	TotalParts      int32  `gorm:"column:TotalParts" json:"TotalParts"`
	Type            bool   `gorm:"column:Type" json:"StatusType"`
	IsCompleted     bool   `gorm:"column:IsCompleted" json:"IsCompleted"`
	RequestStatus   int32  `gorm:"column:RequestStatus" json:"RequestStatus"`
	UploadRequest   []byte `gorm:"column:UploadRequest" json:"UploadRequest"`
	DownloadRequest []byte `gorm:"column:DownloadRequest" json:"DownloadRequest"`
}

func (FileStatus) TableName() string {
	return "filestatus"
}
