package repo

import (
	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/repo/dto"
)

// UISettings repoUISettings interface
type UISettings interface {
	Get(key string) (value string, err error)
	Put(key, value string) error
	Delete(key string) error
}

type repoUISettings struct {
	*repository
}

// Get
func (r *repoUISettings) Get(key string) (value string, err error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	row := new(dto.UISettings)
	err = r.db.Where("Key = ?", key).First(row).Error
	if row == nil {
		err = domain.ErrDoesNotExists
	}
	value = row.Value
	return
}

// Put
func (r *repoUISettings) Put(key string, value string) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	row := dto.UISettings{}

	r.db.Where("Key = ?", key).First(&row)

	row.Key = key
	row.Value = value

	return r.db.Save(row).Error
}

// Delete
func (r *repoUISettings) Delete(key string) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	err := r.db.Where("Key = ?", key).Delete(dto.UISettings{}).Error

	return err
}
