package dto

type FilesStatus struct {
	MessageID       int64  `gorm:"primary_key;column:MessageID" json:"MessageID"`
	FileID          int64  `gorm:"column:FileID" json:"FileID"`
	ClusterID       int32  `gorm:"column:ClusterID" json:"ClusterID"`
	AccessHash      int64  `gorm:"column:AccessHash" json:"AccessHash"`
	Version         int32  `gorm:"column:Version" json:"Version"`
	FilePath        string `gorm:"column:FilePath" json:"FilePath"`
	TotalSize       int64  `gorm:"column:TotalSize" json:"TotalSize"`
	PartList        []byte `gorm:"column:PartList" json:"PartList"`
	TotalParts      int64  `gorm:"column:TotalParts" json:"TotalParts"`
	Type            int32  `gorm:"column:Type" json:"StatusType"`
	IsCompleted     bool   `gorm:"column:IsCompleted" json:"IsCompleted"`
	RequestStatus   int32  `gorm:"column:RequestStatus" json:"RequestStatus"`
	UploadRequest   []byte `gorm:"column:UploadRequest" json:"UploadRequest"`
	DownloadRequest []byte `gorm:"column:DownloadRequest" json:"DownloadRequest"`
	ThumbFileID     int64  `gorm:"column:ThumbFileID" json:"ThumbFileID"`
	ThumbFilePath   string `gorm:"column:ThumbFilePath" json:"ThumbFilePath"`
	ThumbPosition   int64  `gorm:"column:ThumbPosition" json:"ThumbPosition"`
	ThumbTotalSize  int64  `gorm:"column:ThumbTotalSize" json:"ThumbTotalSize"`
	ThumbPartNo     int32  `gorm:"column:ThumbPartNo" json:"ThumbPartNo"`
	ThumbTotalParts int32  `gorm:"column:ThumbTotalParts" json:"ThumbTotalParts"`
}

func (FilesStatus) TableName() string {
	return "files_status"
}
