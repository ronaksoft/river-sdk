package repo

import (
	"github.com/allegro/bigcache"
	"sync"
	"time"

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
	lCache        *bigcache.BigCache

	Dialogs         *repoDialogs
	Messages        *repoMessages
	PendingMessages *repoMessagesPending
	MessagesExtra   *repoMessagesExtra
	System          *repoSystem
	Users           *repoUsers
	UISettings      *repoUISettings
	Groups          *repoGroups
	Files           *repoFiles
)

// Context container of repo
type Context struct {
	DBDialect string
	DBPath    string
}

type repository struct {
	db *gorm.DB
	mx sync.Mutex
}

// create tables
func (r *repository) initDB() error {
	// WARNING: AutoMigrate will ONLY create tables, missing columns and missing indexes,
	// and WON’T change existing column’s type or delete unused columns to protect your data.
	repoLastError = r.db.AutoMigrate(
		dto.Dialogs{},
		dto.Messages{},
		dto.MessagesPending{},
		dto.MessagesExtra{},
		dto.System{},
		dto.Users{},
		dto.UISettings{},
		dto.Groups{},
		dto.GroupsParticipants{},
		dto.FilesStatus{},
		dto.Files{},
		dto.UsersPhoto{},
	).Error

	return repoLastError
}

// InitRepo initialize repo singleton
func InitRepo(dialect, dbPath string) error {
	if ctx == nil {
		singleton.Lock()
		lcConfig := bigcache.DefaultConfig(time.Second*360)
		lcConfig.CleanWindow = time.Second * 30
		lcConfig.HardMaxCacheSize = 128
		lCache, _ = bigcache.NewBigCache(lcConfig)
		repoLastError = repoSetDB(dialect, dbPath)
		ctx = &Context{
			DBDialect: dialect,
			DBPath:    dbPath,
		}
		Dialogs = &repoDialogs{repository: r}
		Messages = &repoMessages{repository: r}
		PendingMessages = &repoMessagesPending{repository: r}
		MessagesExtra = &repoMessagesExtra{repository: r}
		System = &repoSystem{repository: r}
		Users = &repoUsers{repository: r}
		UISettings = &repoUISettings{repository: r}
		Groups = &repoGroups{repository: r}
		Files = &repoFiles{repository: r}
		singleton.Unlock()
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

// DropAndCreateTable remove and create elated dto object table
func DropAndCreateTable(dtoTable interface{}) error {
	err := r.db.DropTable(dtoTable).Error
	if err != nil {
		logs.Warn("Context::DropAndCreateTable() failed to DROP", zap.Error(err))
	}
	err = r.db.AutoMigrate(dtoTable).Error
	if err != nil {
		logs.Error("Context::DropAndCreateTable() failed to AutoMigrate", zap.Error(err))
	}
	return err
}

// ReInitiateDatabase runs auto migrate
func ReInitiateDatabase() error {
	err := r.db.DropTableIfExists(
		dto.Dialogs{},
		dto.Messages{},
		dto.MessagesPending{},
		dto.MessagesExtra{},
		dto.System{},
		dto.Users{},
		dto.Groups{},
		dto.GroupsParticipants{},
		dto.FilesStatus{},
		dto.Files{},
		dto.UsersPhoto{},
		// dto.UISettings{}, //do not remove UISettings on logout
	).Error

	if err != nil {
		return err
	}

	err = r.initDB()

	return err
}

// Close underlying DB connection
func Close() error {
	repoLastError = r.db.Close()
	r = nil
	ctx = nil
	return repoLastError
}
