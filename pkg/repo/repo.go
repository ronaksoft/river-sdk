package repo

import (
	"fmt"
	"github.com/allegro/bigcache"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/mapping"
	"github.com/dgraph-io/badger/options"
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
	badger     *badger.DB
	bunt       *buntdb.DB
	msgSearch  bleve.Index
	peerSearch bleve.Index
}

// InitRepo initialize repo singleton
func InitRepo(dbPath string, lowMemory bool) error {
	if ctx == nil {
		singleton.Lock()
		lcConfig := bigcache.DefaultConfig(time.Second * 360)
		lcConfig.CleanWindow = time.Second * 30
		lcConfig.MaxEntrySize = 1024
		lcConfig.MaxEntriesInWindow = 10000
		if lowMemory {
			lcConfig.HardMaxCacheSize = 8
		} else {
			lcConfig.HardMaxCacheSize = 64
		}

		lCache, _ = bigcache.NewBigCache(lcConfig)
		repoLastError = repoSetDB(dbPath, lowMemory)
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

func repoSetDB(dbPath string, lowMemory bool) error {
	r = new(repository)

	_ = os.MkdirAll(dbPath, os.ModePerm)
	// Initialize BadgerDB
	_ = os.MkdirAll(fmt.Sprintf("%s/badger", strings.TrimRight(dbPath, "/")), os.ModePerm)
	badgerOpts := badger.DefaultOptions(fmt.Sprintf("%s/badger", strings.TrimRight(dbPath, "/"))).
		WithLogger(nil)
	if lowMemory {
		badgerOpts = badgerOpts.
			WithTableLoadingMode(options.FileIO).
			WithValueLogLoadingMode(options.FileIO).
			WithValueLogFileSize(1 << 24) // 16MB
	} else {
		badgerOpts = badgerOpts.
			WithTableLoadingMode(options.LoadToRAM).
			WithValueLogLoadingMode(options.FileIO)
	}
	r.badger, repoLastError = badger.Open(badgerOpts)
	if repoLastError != nil {
		logs.Info("Context::repoSetDB()->badger Open()",
			zap.String("Error", repoLastError.Error()),
		)
		return repoLastError
	}

	// Initialize BuntDB Indexer
	_ = os.MkdirAll(fmt.Sprintf("%s/bunty", strings.TrimRight(dbPath, "/")), os.ModePerm)
	r.bunt, repoLastError = buntdb.Open(fmt.Sprintf("%s/bunty/dialogs.db", strings.TrimRight(dbPath, "/")))
	if repoLastError != nil {
		logs.Info("Context::repoSetDB()->bunt Open()",
			zap.String("Error", repoLastError.Error()),
		)
		return repoLastError
	}
	_ = r.bunt.Update(func(tx *buntdb.Tx) error {
		return tx.CreateIndex(indexDialogs, fmt.Sprintf("%s.*", prefixDialogs), buntdb.IndexBinary)
	})

	// Initialize Search
	// 1. Messages Search
	searchDbPath := fmt.Sprintf("%s/searchdb/msg", strings.TrimRight(dbPath, "/"))
	r.msgSearch, repoLastError = bleve.Open(searchDbPath)
	if repoLastError == bleve.ErrorIndexPathDoesNotExist {
		repoLastError = nil
		// create a mapping
		indexMapping, err := indexMapForMessages()
		if err != nil {
			logs.Fatal("BuildIndexMapping For Messages", zap.Error(err))
		}
		r.msgSearch, err = bleve.New(searchDbPath, indexMapping)
		if err != nil {
			logs.Fatal("New SearchIndex for Messages", zap.Error(err))
		}
	} else if repoLastError != nil {
		logs.Fatal("Error Opening SearchIndex for Messages", zap.Error(repoLastError))
	}

	// 2. Peer Search
	peerDbSearch := fmt.Sprintf("%s/searchdb/peer", strings.TrimRight(dbPath, "/"))
	r.peerSearch, repoLastError = bleve.Open(peerDbSearch)
	if repoLastError == bleve.ErrorIndexPathDoesNotExist {
		repoLastError = nil
		// create a mapping
		indexMapping, err := indexMapForPeers()
		if err != nil {
			logs.Fatal("BuildIndexMapping For Peers", zap.Error(err))
		}
		r.peerSearch, err = bleve.New(peerDbSearch, indexMapping)
		if err != nil {
			logs.Fatal("New SearchIndex for Peers", zap.Error(err))
		}
	} else if repoLastError != nil {
		logs.Fatal("Error Opening SearchIndex for Peers", zap.Error(repoLastError))
	}

	return repoLastError
}

func indexMapForMessages() (mapping.IndexMapping, error) {
	// a generic reusable mapping for english text
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = en.AnalyzerName
	textFieldMapping.Store = false
	textFieldMapping.IncludeTermVectors = false
	textFieldMapping.DocValues = false
	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = keyword.Name

	// Message
	messageMapping := bleve.NewDocumentStaticMapping()
	messageMapping.AddFieldMappingsAt("Body", textFieldMapping)
	messageMapping.AddFieldMappingsAt("PeerID", keywordFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("msg", messageMapping)

	indexMapping.TypeField = "type"
	indexMapping.DefaultAnalyzer = en.AnalyzerName

	return indexMapping, nil
}

func indexMapForPeers() (mapping.IndexMapping, error) {
	// a generic reusable mapping for english text
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Store = false
	textFieldMapping.IncludeTermVectors = false
	textFieldMapping.DocValues = false
	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = keyword.Name

	// User
	userMapping := bleve.NewDocumentStaticMapping()
	userMapping.AddFieldMappingsAt("FirstName", textFieldMapping)
	userMapping.AddFieldMappingsAt("LastName", textFieldMapping)
	userMapping.AddFieldMappingsAt("Username", keywordFieldMapping)
	userMapping.AddFieldMappingsAt("Phone", keywordFieldMapping)

	// GroupSearch
	groupMapping := bleve.NewDocumentStaticMapping()
	groupMapping.AddFieldMappingsAt("Title", textFieldMapping)

	// Contact
	contactMapping := bleve.NewDocumentStaticMapping()
	contactMapping.AddFieldMappingsAt("FirstName", textFieldMapping)
	contactMapping.AddFieldMappingsAt("LastName", textFieldMapping)
	contactMapping.AddFieldMappingsAt("Username", keywordFieldMapping)
	contactMapping.AddFieldMappingsAt("Phone", keywordFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("user", userMapping)
	indexMapping.AddDocumentMapping("group", groupMapping)
	indexMapping.AddDocumentMapping("contact", contactMapping)

	indexMapping.TypeField = "type"
	indexMapping.DefaultAnalyzer = en.AnalyzerName

	return indexMapping, nil
}

// Close underlying DB connection
func Close() error {
	logs.Debug("Repo Stopping")

	_ = r.bunt.Close()
	_ = r.badger.Close()

	_ = r.msgSearch.Close()
	_ = r.peerSearch.Close()
	r = nil
	ctx = nil
	logs.Debug("Repo Stopped")
	return repoLastError
}

func DropAll() {
	_ = r.badger.DropAll()
	_ = r.bunt.Shrink()
}

func GC() {
	_ = r.bunt.Shrink()
	_ = r.badger.RunValueLogGC(0.5)
}

func DbSize() (int64, int64) {
	return r.badger.Size()
}

func TableInfo() []badger.TableInfo {
	r.badger.Size()
	return r.badger.Tables(true)
}

func WarnOnErr(guideTxt string, err error, fields ...zap.Field) {
	if err != nil {
		fields = append(fields, zap.Error(err))
		logs.Warn(guideTxt, fields...)
	}
}
