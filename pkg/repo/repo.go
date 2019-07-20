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
	"os"
	"strings"
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"github.com/dgraph-io/badger"
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
	Groups          *repoGroups
	Files           *repoFiles
)

// Context container of repo
type Context struct {
	DBPath string
}

type repository struct {
	badger      *badger.DB
	bunt        *buntdb.DB
	searchIndex bleve.Index
	// mx          sync.Mutex
}

// create tables
func (r *repository) initDB() error {
	return repoLastError
}

// InitRepo initialize repo singleton
func InitRepo(dbPath string) error {
	if ctx == nil {
		singleton.Lock()
		lcConfig := bigcache.DefaultConfig(time.Second * 360)
		lcConfig.CleanWindow = time.Second * 30
		lcConfig.HardMaxCacheSize = 128
		lCache, _ = bigcache.NewBigCache(lcConfig)
		repoLastError = repoSetDB(dbPath)
		ctx = &Context{
			DBPath: dbPath,
		}
		Dialogs = &repoDialogs{repository: r}
		Messages = &repoMessages{repository: r}
		PendingMessages = &repoMessagesPending{repository: r}
		MessagesExtra = &repoMessagesExtra{repository: r}
		System = &repoSystem{repository: r}
		Users = &repoUsers{repository: r}
		Groups = &repoGroups{repository: r}
		Files = &repoFiles{repository: r}
		singleton.Unlock()
	}
	return repoLastError
}

func repoSetDB(dbPath string) error {
	r = new(repository)
	_ = os.MkdirAll(fmt.Sprintf("%s/badger", strings.TrimRight(dbPath, "/")), os.ModePerm)
	r.badger, repoLastError = badger.Open(badger.DefaultOptions(fmt.Sprintf("%s/badger", strings.TrimRight(dbPath, "/"))).WithLogger(nil))
	if repoLastError != nil {
		logs.Debug("Context::repoSetDB()->badger Open()",
			zap.String("Error", repoLastError.Error()),
		)
		return repoLastError
	}
	_ = os.MkdirAll(fmt.Sprintf("%s/bunty", strings.TrimRight(dbPath, "/")), os.ModePerm)
	r.bunt, repoLastError = buntdb.Open(fmt.Sprintf("%s/bunty/dialogs.db", strings.TrimRight(dbPath, "/")))
	if repoLastError != nil {
		logs.Debug("Context::repoSetDB()->bunt Open()",
			zap.String("Error", repoLastError.Error()),
		)
		return repoLastError
	}
	_ = r.bunt.Update(func(tx *buntdb.Tx) error {
		return tx.CreateIndex(indexDialogs, fmt.Sprintf("%s.*", prefixDialogs), buntdb.IndexBinary)
	})

	// _ = os.MkdirAll(fmt.Sprintf("%s", strings.TrimRight(dbPath, "/")), os.ModePerm)
	r.searchIndex, repoLastError = bleve.Open(fmt.Sprintf("%s/bleve/", strings.TrimRight(dbPath, "/")))
	if repoLastError == bleve.ErrorIndexPathDoesNotExist {
		repoLastError = nil
		// create a mapping
		indexMapping, err := buildIndexMapping()
		if err != nil {
			logs.Fatal("Build Index", zap.Error(err))
		}
		r.searchIndex, err = bleve.New(fmt.Sprintf("%s/bleve", strings.TrimRight(dbPath, "/")), indexMapping)
		if err != nil {
			logs.Fatal("SearchIndex", zap.Error(err))
		}
	} else if repoLastError != nil {
		logs.Fatal("Another", zap.Error(repoLastError))
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
	err := r.badger.DropAll()
	err = r.initDB()

	return err
}

// Close underlying DB connection
func Close() error {
	logs.Debug("Repo Stopping")

	_ = r.bunt.Close()
	_ = r.badger.Close()
	_ = r.searchIndex.Close()
	r = nil
	ctx = nil
	logs.Debug("Repo Stopped")
	return repoLastError
}
