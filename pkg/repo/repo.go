package repo

import (
	"fmt"
	"github.com/allegro/bigcache"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/blevex/detectlang"
	"github.com/tidwall/buntdb"
	"log"
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"
	"github.com/dgraph-io/badger"
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
	db          *gorm.DB
	badger      *badger.DB
	bunt        *buntdb.DB
	searchIndex bleve.Index
	mx          sync.Mutex
}

// create tables
func (r *repository) initDB() error {
	// WARNING: AutoMigrate will ONLY create tables, missing columns and missing indexes,
	// and WON’T change existing column’s type or delete unused columns to protect your data.
	repoLastError = r.db.AutoMigrate(
		dto.MessagesPending{},
		dto.MessagesExtra{},
		dto.UISettings{},
		dto.FilesStatus{},
		dto.Files{},
	).Error

	return repoLastError
}

// InitRepo initialize repo singleton
func InitRepo(dialect, dbPath string) error {
	if ctx == nil {
		singleton.Lock()
		lcConfig := bigcache.DefaultConfig(time.Second * 360)
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
	r.badger, repoLastError = badger.Open(badger.DefaultOptions(dbPath).WithLogger(nil))
	if repoLastError != nil {
		logs.Debug("Context::repoSetDB()->badger Open()",
			zap.String("Error", repoLastError.Error()),
		)
		return repoLastError
	}
	r.bunt, repoLastError = buntdb.Open(dbPath)
	if repoLastError != nil {
		logs.Debug("Context::repoSetDB()->bunt Open()",
			zap.String("Error", repoLastError.Error()),
		)
		return repoLastError
	}
	_ = r.bunt.Update(func(tx *buntdb.Tx) error {
		return tx.CreateIndex(indexDialogs, fmt.Sprintf("%s.*", prefixDialogs), buntdb.IndexBinary)
	})

	r.searchIndex, repoLastError = bleve.Open(dbPath)
	if repoLastError == bleve.ErrorIndexPathDoesNotExist {
		// create a mapping
		indexMapping, err := buildIndexMapping()
		if err != nil {
			log.Fatal(err)
		}
		r.searchIndex, err = bleve.New(dbPath, indexMapping)
		if err != nil {
			log.Fatal(err)
		}
	} else if repoLastError != nil {
		log.Fatal(repoLastError)
	}

	return r.initDB()
}

func buildIndexMapping() (mapping.IndexMapping, error) {
	// a generic reusable mapping for english text
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = detectlang.AnalyzerName
	textFieldMapping.Store = false
	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = keyword.Name

	// Message
	messageMapping := bleve.NewDocumentStaticMapping()
	messageMapping.AddFieldMappingsAt("Body", textFieldMapping)
	messageMapping.AddFieldMappingsAt("PeerID", keywordFieldMapping)

	// User
	userMapping := bleve.NewDocumentStaticMapping()
	userMapping.AddFieldMappingsAt("FirstName", textFieldMapping)
	userMapping.AddFieldMappingsAt("LastName", textFieldMapping)
	userMapping.AddFieldMappingsAt("Username", keywordFieldMapping)
	userMapping.AddFieldMappingsAt("Phone", keywordFieldMapping)

	// Group
	groupMapping := bleve.NewDocumentStaticMapping()
	groupMapping.AddFieldMappingsAt("Title", textFieldMapping)

	// Contact
	contactMapping := bleve.NewDocumentStaticMapping()
	contactMapping.AddFieldMappingsAt("FirstName", textFieldMapping)
	contactMapping.AddFieldMappingsAt("LastName", textFieldMapping)
	contactMapping.AddFieldMappingsAt("Username", keywordFieldMapping)
	contactMapping.AddFieldMappingsAt("Phone", keywordFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("msg", messageMapping)
	indexMapping.AddDocumentMapping("user", userMapping)
	indexMapping.AddDocumentMapping("group", groupMapping)
	indexMapping.AddDocumentMapping("contact", contactMapping)

	indexMapping.TypeField = "type"
	indexMapping.DefaultAnalyzer = en.AnalyzerName

	return indexMapping, nil
}

// ReInitiateDatabase runs auto migrate
func ReInitiateDatabase() error {
	err := r.db.DropTableIfExists(
		dto.MessagesPending{},
		dto.MessagesExtra{},
		dto.FilesStatus{},
		dto.Files{},
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
	logs.Debug("Repo Stopping")

	_ = r.bunt.Close()
	_ = r.badger.Close()
	repoLastError = r.db.Close()
	r = nil
	ctx = nil
	logs.Debug("Repo Stopped")
	return repoLastError
}
