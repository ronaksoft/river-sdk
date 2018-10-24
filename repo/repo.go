package repo

import (
	"encoding/json"
	"sync"

	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/repo/dto"
	"go.uber.org/zap"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

var (
	ctx           *Context
	r             *repository
	singletone    sync.Mutex
	repoLastError error
)

type Context struct {
	Dialogs         RepoDialogs
	Messages        RepoMessages
	PendingMessages RepoPendingMessages
	System          RepoSystem
	Users           RepoUsers
	UISettings      RepoUISettings
}

type repository struct {
	db *gorm.DB
	mx sync.Mutex
}

// Ctx return repository context
func Ctx() *Context {
	if ctx == nil {
		panic("repo not initialized !")
	}
	return ctx
}

func InitRepo(dialect, dbPath string) error {
	if ctx == nil {
		singletone.Lock()
		defer singletone.Unlock()
		if ctx == nil {
			repoLastError = repoSetDB(dialect, dbPath)
			ctx = &Context{
				Dialogs:         &repoDialogs{repository: r},
				Messages:        &repoMessages{repository: r},
				PendingMessages: &repoPendingMessages{repository: r},
				System:          &repoSystem{repository: r},
				Users:           &repoUsers{repository: r},
				UISettings:      &repoUISettings{repository: r},
			}
		}
	}

	// singletone.Do(func() {
	// 	repoLastError = repoSetDB(dbPath, dbID)
	// 	ctx = &Context{
	// 		Dialogs:         &repoDialogs{repository: r},
	// 		Messages:        &repoMessages{repository: r},
	// 		PendingMessages: &repoPendingMessages{repository: r},
	// 		System:          &repoSystem{repository: r},
	// 		Users:           &repoUsers{repository: r},
	// 	}
	// })

	return repoLastError
}

func repoSetDB(dialect, dbPath string) error {
	r = new(repository)
	r.db, repoLastError = gorm.Open(dialect, dbPath)
	if repoLastError != nil {
		return repoLastError
	}

	return r.initDB()

}

func (c *Context) LogMode(enable bool) {

	if r != nil {
		r.db.LogMode(enable)
	}
}

func (c *Context) Close() error {
	repoLastError = r.db.Close()
	r = nil
	ctx = nil
	return repoLastError
}

func (c *Context) DropAndCreateTable(dtoTable interface{}) error {
	err := r.db.DropTable(dtoTable).Error
	if err != nil {
		log.LOG.Debug("repo::DropAndCreateTable() failed to DROP",
			zap.String("error", err.Error()),
		)
	}
	err = r.db.AutoMigrate(dtoTable).Error
	if err != nil {
		log.LOG.Debug("repo::DropAndCreateTable() failed to AutoMigrate",
			zap.String("error", err.Error()),
		)
	}
	return err
}

func (c *Context) ReinitiateDatabase() error {
	err := r.db.DropTableIfExists(
		dto.Dialogs{},
		dto.Messages{},
		dto.PendingMessages{},
		dto.System{},
		dto.Users{},
		//dto.UISettings{}, //do not remove UISettings on logout
	).Error

	if err != nil {
		log.LOG.Debug("repo::ReinitiateDatabase()",
			zap.String("error", err.Error()),
		)
	}

	err = r.initDB()

	return err
}

// create tables
func (r *repository) initDB() error {

	// WARNING: AutoMigrate will ONLY create tables, missing columns and missing indexes,
	// and WON’T change existing column’s type or delete unused columns to protect your data.
	repoLastError = r.db.AutoMigrate(
		dto.Dialogs{},
		dto.Messages{},
		dto.PendingMessages{},
		dto.System{},
		dto.Users{},
		dto.UISettings{},
	).Error

	return repoLastError
}

func (r *repository) Map(from interface{}, to interface{}) error {

	buff, err := json.Marshal(from)
	json.Unmarshal(buff, to)
	return err
}

// func (r *repository) IsNew(dto interface{}) bool {
// 	return r.db.Model(dto).NewRecord(dto)
// }

// func (r *repository) Update(dto interface{}) error {
// 	return r.db.Model(dto).Save(dto).Error
// }

// func (r *repository) Insert(dto interface{}) error {
// 	return r.db.Model(dto).Create(dto).Error
// }

// func (r *repository) Delete(dto interface{}) error {
// 	return r.db.Model(dto).Delete(dto).Error
// }
