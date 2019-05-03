package repo

import (
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

	Dialogs         = &repoDialogs{repository: r}
	Messages        = &repoMessages{repository: r}
	PendingMessages = &repoMessagesPending{repository: r}
	MessagesExtra   = &repoMessagesExtra{repository: r}
	MessageHoles    = &repoMessagesHole{repository: r}
	System          = &repoSystem{repository: r}
	Users           = &repoUsers{repository: r}
	UISettings      = &repoUISettings{repository: r}
	Groups          = &repoGroups{repository: r}
	Files           = &repoFiles{repository: r}
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
		dto.MessagesHole{},
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
		defer singleton.Unlock()
		repoLastError = repoSetDB(dialect, dbPath)
		ctx = &Context{
			DBDialect: dialect,
			DBPath:    dbPath,
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

// DropAndCreateTable remove and create elated dto object table
func DropAndCreateTable(dtoTable interface{}) error {
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
		dto.MessagesHole{},
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

// // Exec execute raw query
// func (c *Context) Exec(qry string) error {
// 	return r.db.Exec(qry).Error
// }

// // Map basic mapper don't use this define mapper for each dto
// func (r *repository) Map(from interface{}, to interface{}) error {
// 	buff, err := json.Marshal(from)
// 	if err != nil {
// 		return err
// 	}
// 	err = json.Unmarshal(buff, to)
// 	return err
// }

// // Exec execute raw query
// func (r *repository) Exec(qry string) error {
// 	return r.db.Exec(qry).Error
// }


// // LogMode set query logger if true prints all executed queries
// func (c *Context) LogMode(enable bool) {
// 	if r != nil {
// 		r.db.LogMode(enable)
// 	}
// }

