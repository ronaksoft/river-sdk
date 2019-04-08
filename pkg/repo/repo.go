package repo

import (
	"encoding/json"
	"sync"

	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"

	_ "github.com/mattn/go-sqlite3"
)

var (
	ctx           *Context
	r             *repository
	singleton     sync.Mutex
	repoLastError error
)

// Context container of repo
type Context struct {
	DBDialect       string
	DBPath          string
	Dialogs         Dialogs
	Messages        Messages
	PendingMessages PendingMessages
	System          System
	Users           Users
	UISettings      UISettings
	Groups          Groups
	MessageHoles    MessageHoles
	Files           Files
	ScrollStatus	ScrollStatus
}

type repository struct {
	db *gorm.DB
	mx sync.Mutex
}

// Ctx return repository context
func Ctx() *Context {
	if ctx == nil {
		panic("Context::Ctx() repo not initialized !")
	}
	return ctx
}

// InitRepo initialize repo singleton
func InitRepo(dialect, dbPath string) error {
	if ctx == nil {
		singleton.Lock()
		defer singleton.Unlock()
		if ctx == nil {
			repoLastError = repoSetDB(dialect, dbPath)
			ctx = &Context{
				DBDialect:       dialect,
				DBPath:          dbPath,
				Dialogs:         &repoDialogs{repository: r},
				Messages:        &repoMessages{repository: r},
				PendingMessages: &repoPendingMessages{repository: r},
				System:          &repoSystem{repository: r},
				Users:           &repoUsers{repository: r},
				UISettings:      &repoUISettings{repository: r},
				Groups:          &repoGroups{repository: r},
				MessageHoles:    &repoMessageHoles{repository: r},
				Files:           &repoFiles{repository: r},
			}
		}
	}

	return repoLastError
}

func repoSetDB(dialect, dbPath string) error {
	r = new(repository)
	r.db, repoLastError = gorm.Open(dialect, dbPath)
	if repoLastError != nil {
		logs.Debug("Context::repoSetDB()->gorm.Open()",
			zap.String("Error", repoLastError.Error()),
		)
		return repoLastError
	}

	return r.initDB()
}

// LogMode set query logger if true prints all executed queries
func (c *Context) LogMode(enable bool) {

	if r != nil {
		r.db.LogMode(enable)
	}
}

// Close underlying DB connection
func (c *Context) Close() error {
	repoLastError = r.db.Close()
	r = nil
	ctx = nil
	return repoLastError
}

// DropAndCreateTable remove and create elated dto object table
func (c *Context) DropAndCreateTable(dtoTable interface{}) error {
	err := r.db.DropTable(dtoTable).Error
	if err != nil {
		logs.Error("Context::DropAndCreateTable() failed to DROP", zap.Error(err))
	}
	err = r.db.AutoMigrate(dtoTable).Error
	if err != nil {
		logs.Error("Context::DropAndCreateTable() failed to AutoMigrate", zap.Error(err))
	}
	return err
}

// ReInitiateDatabase runs auto migrate
func (c *Context) ReInitiateDatabase() error {
	err := r.db.DropTableIfExists(
		dto.Dialogs{},
		dto.Messages{},
		dto.PendingMessages{},
		dto.System{},
		dto.Users{},
		dto.Groups{},
		dto.GroupParticipants{},
		dto.MessageHoles{},
		dto.FileStatus{},
		dto.Files{},
		dto.UserPhotos{},
		// dto.UISettings{}, //do not remove UISettings on logout
	).Error

	if err != nil {
		logs.Error("Context::ReInitiateDatabase()->DropTableIfExists()", zap.Error(err))
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
		dto.Groups{},
		dto.GroupParticipants{},
		dto.MessageHoles{},
		dto.FileStatus{},
		dto.Files{},
		dto.UserPhotos{},
	).Error

	return repoLastError
}

// Exec execute raw query
func (c *Context) Exec(qry string) error {
	return r.db.Exec(qry).Error
}

// Map basic mapper don't use this define mapper for each dto
func (r *repository) Map(from interface{}, to interface{}) error {
	buff, err := json.Marshal(from)
	if err != nil {
		return err
	}
	err = json.Unmarshal(buff, to)
	return err
}

// Exec execute raw query
func (r *repository) Exec(qry string) error {
	return r.db.Exec(qry).Error
}
