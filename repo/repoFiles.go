package repo

import (
	"git.ronaksoftware.com/ronak/riversdk/repo/dto"
)

type RepoFiles interface {
	SaveFileStatus(fileID int64, filePath string, position, totalSize int64, partNo, totalParts int32) (err error)
}

type repoFiles struct {
	*repository
}

// SaveFileStatus
func (r *repoFiles) SaveFileStatus(fileID int64, filePath string, position, totalSize int64, partNo, totalParts int32) (err error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	dto := new(dto.FileStatus)
	err = r.db.Where("FileID = ?", fileID).First(dto).Error
	if dto == nil {
		return r.db.Create(dto).Error
	}
	return r.db.Save(dto).Error
}
