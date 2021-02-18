package repo

import (
	"fmt"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/mapping"
	"github.com/dgraph-io/badger/v2/options"
	"github.com/pkg/errors"
	"github.com/tidwall/buntdb"
	"os"
	"strings"
	"sync"
	"time"

	"git.ronaksoft.com/river/sdk/internal/logs"
	"github.com/dgraph-io/badger/v2"
	"go.uber.org/zap"
)

var (
	ctx       *Context
	r         *repository
	singleton sync.Mutex

	Account         *repoAccount
	Dialogs         *repoDialogs
	Messages        *repoMessages
	PendingMessages *repoMessagesPending
	MessagesExtra   *repoMessagesExtra
	System          *repoSystem
	Users           *repoUsers
	Gifs            *repoGifs
	Groups          *repoGroups
	Files           *repoFiles
	Labels          *repoLabels
	TopPeers        *repoTopPeers
	Wallpapers      *repoWallpapers
	RecentSearches  *repoRecentSearches
	Teams           *repoTeams
	Reactions       *repoReactions
)

// Context container of repo
type Context struct {
	DBPath string
}

type repository struct {
	badger     *badger.DB
	selfUserID int64
	bunt       *buntdb.DB
	msgSearch  bleve.Index
	peerSearch bleve.Index
}

func MustInitRepo(dbPath string, lowMemory bool) {
	err := InitRepo(dbPath, lowMemory)
	if err != nil {
		panic(err)
	}
}

// InitRepo initialize repo singleton
func InitRepo(dbPath string, lowMemory bool) error {
	if ctx == nil {
		singleton.Lock()
		err := repoSetDB(dbPath, lowMemory)
		if err != nil {
			return err
		}

		ctx = &Context{
			DBPath: dbPath,
		}
		Account = &repoAccount{repository: r}
		Dialogs = &repoDialogs{repository: r}
		Messages = &repoMessages{repository: r}
		PendingMessages = &repoMessagesPending{repository: r}
		MessagesExtra = &repoMessagesExtra{repository: r}
		System = &repoSystem{repository: r}
		Users = &repoUsers{repository: r}
		Groups = &repoGroups{repository: r}
		Gifs = &repoGifs{repository: r}
		Files = &repoFiles{repository: r}
		Labels = &repoLabels{repository: r}
		TopPeers = &repoTopPeers{repository: r}
		Wallpapers = &repoWallpapers{repository: r}
		RecentSearches = &repoRecentSearches{repository: r}
		Teams = &repoTeams{repository: r}
		Reactions = &repoReactions{repository: r}
		singleton.Unlock()
	}
	return nil
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
			WithNumMemtables(2).
			WithNumLevelZeroTables(2).
			WithNumLevelZeroTablesStall(4).
			WithMaxTableSize(1 << 22).     // 4MB
			WithValueLogFileSize(1 << 22). // 4MB
			WithBypassLockGuard(true)
	} else {
		badgerOpts = badgerOpts.
			WithTableLoadingMode(options.FileIO).
			WithValueLogLoadingMode(options.FileIO).
			WithBypassLockGuard(true)

	}
	if badgerDB, err := badger.Open(badgerOpts); err != nil {
		return errors.Wrap(err, "Badger")
	} else {
		r.badger = badgerDB
	}

	// Initialize BuntDB Indexer
	_ = os.MkdirAll(fmt.Sprintf("%s/bunty", strings.TrimRight(dbPath, "/")), os.ModePerm)
	if buntIndex, err := buntdb.Open(fmt.Sprintf("%s/bunty/dialogs.db", strings.TrimRight(dbPath, "/"))); err != nil {
		logs.Fatal("Context::repoSetDB()->bunt Open()", zap.Error(err))
	} else {
		r.bunt = buntIndex
	}

	_ = r.bunt.Update(func(tx *buntdb.Tx) error {
		_ = tx.CreateIndex(indexDialogs, fmt.Sprintf("%s.*", prefixDialogs), buntdb.IndexBinary)
		_ = tx.CreateIndex(indexTopPeersUser, fmt.Sprintf("%s.*", indexTopPeersUser), buntdb.IndexFloat)
		_ = tx.CreateIndex(indexTopPeersGroup, fmt.Sprintf("%s.*", indexTopPeersGroup), buntdb.IndexFloat)
		_ = tx.CreateIndex(indexTopPeersForward, fmt.Sprintf("%s.*", indexTopPeersForward), buntdb.IndexFloat)
		_ = tx.CreateIndex(indexTopPeersBotMessage, fmt.Sprintf("%s.*", indexTopPeersBotMessage), buntdb.IndexFloat)
		_ = tx.CreateIndex(indexTopPeersBotInline, fmt.Sprintf("%s.*", indexTopPeersBotInline), buntdb.IndexFloat)
		_ = tx.CreateIndex(indexGif, fmt.Sprintf("%s.*", prefixGif), buntdb.IndexBinary)

		return nil
	})

	// Initialize Search
	go func() {
		// 1. Messages Search
		_ = domain.Try(10, time.Millisecond*100, func() error {
			searchDbPath := fmt.Sprintf("%s/searchdb/msg", strings.TrimRight(dbPath, "/"))
			if msgSearch, err := bleve.Open(searchDbPath); err != nil {
				switch err {
				case bleve.ErrorIndexPathDoesNotExist:
					// create a mapping
					r.msgSearch, err = bleve.New(searchDbPath, indexMapForMessages())
					if err != nil {
						logs.Warn("Error On Open Search(Message)[New]", zap.Error(err))
						_ = os.RemoveAll(searchDbPath)
						return err
					}
				default:
					logs.Warn("Error On Open Search(Message)[Default]", zap.Error(err))
					_ = os.RemoveAll(searchDbPath)
					return err
				}
			} else {
				r.msgSearch = msgSearch
				logs.Info("Repo initialized Message Index successfully.")
			}
			return nil
		})
	}()
	go func() {
		// 2. Peer Search
		_ = domain.Try(10, 100*time.Millisecond, func() error {
			peerDbPath := fmt.Sprintf("%s/searchdb/peer", strings.TrimRight(dbPath, "/"))
			if peerSearch, err := bleve.Open(peerDbPath); err != nil {
				switch err {
				case bleve.ErrorIndexPathDoesNotExist:
					// create a mapping
					r.peerSearch, err = bleve.New(peerDbPath, indexMapForPeers())
					if err != nil {
						logs.Warn("Error On Open Search(Peers)", zap.Error(err))
						_ = os.RemoveAll(peerDbPath)
						return err
					}
				default:
					logs.Warn("Error On Open Search(Peers)", zap.Error(err))
					_ = os.RemoveAll(peerDbPath)
					return err
				}
			} else {
				r.peerSearch = peerSearch
				logs.Info("Repo initialized Peer index successfully.")
			}
			return nil
		})
	}()

	return nil
}

func indexMapForMessages() mapping.IndexMapping {
	// a generic reusable mapping for english text
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = en.AnalyzerName
	textFieldMapping.Store = false
	textFieldMapping.IncludeTermVectors = true
	textFieldMapping.DocValues = false
	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = keyword.Name

	// Message
	messageMapping := bleve.NewDocumentStaticMapping()
	messageMapping.AddFieldMappingsAt("body", textFieldMapping)
	messageMapping.AddFieldMappingsAt("peer_id", keywordFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("msg", messageMapping)

	indexMapping.TypeField = "type"
	indexMapping.DefaultAnalyzer = en.AnalyzerName

	return indexMapping
}

func indexMapForPeers() mapping.IndexMapping {
	// a generic reusable mapping for english text
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Store = false
	textFieldMapping.IncludeTermVectors = true
	textFieldMapping.DocValues = false
	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = keyword.Name

	// User
	userMapping := bleve.NewDocumentStaticMapping()
	userMapping.AddFieldMappingsAt("fn", textFieldMapping)
	userMapping.AddFieldMappingsAt("ln", textFieldMapping)
	userMapping.AddFieldMappingsAt("un", keywordFieldMapping)
	userMapping.AddFieldMappingsAt("phone", keywordFieldMapping)

	// GroupSearch
	groupMapping := bleve.NewDocumentStaticMapping()
	groupMapping.AddFieldMappingsAt("title", textFieldMapping)

	// Contact
	contactMapping := bleve.NewDocumentStaticMapping()
	contactMapping.AddFieldMappingsAt("fn", textFieldMapping)
	contactMapping.AddFieldMappingsAt("ln", textFieldMapping)
	contactMapping.AddFieldMappingsAt("un", keywordFieldMapping)
	contactMapping.AddFieldMappingsAt("phone", keywordFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("user", userMapping)
	indexMapping.AddDocumentMapping("group", groupMapping)
	indexMapping.AddDocumentMapping("contact", contactMapping)

	indexMapping.TypeField = "type"
	indexMapping.DefaultAnalyzer = en.AnalyzerName

	return indexMapping
}

func SetSelfUserID(value int64) {
	r.selfUserID = value
}

func DropAll() {
	SetSelfUserID(0)
	_ = r.bunt.Close()
	_ = r.badger.Close()
	_ = r.msgSearch.Close()
	_ = r.peerSearch.Close()
	for os.RemoveAll(ctx.DBPath) != nil {
		time.Sleep(time.Millisecond * 100)
	}
	ctx = nil
}

func GC() {
	_ = r.bunt.Shrink()
	for r.badger.RunValueLogGC(0.7) == nil {
		logs.Info("Badger ValueLog GC executed")
	}
}

func DbSize() (int64, int64) {
	return r.badger.Size()
}

func badgerUpdate(fn func(txn *badger.Txn) error) (err error) {
	for retry := 100; retry > 0; retry-- {
		err = r.badger.Update(fn)
		switch err {
		case nil:
			return nil
		case badger.ErrConflict:
			logs.Debug("Badger update conflict")
		default:
			logs.Debug("Badger update error", zap.Error(err))
			return
		}
		time.Sleep(time.Duration(domain.RandomInt(10000)) * time.Microsecond)
	}
	return
}

func badgerView(fn func(txn *badger.Txn) error) (err error) {
	for retry := 100; retry > 0; retry-- {
		err = r.badger.View(fn)
		switch err {
		case nil:
			return nil
		case badger.ErrConflict:
		default:
			return
		}
		time.Sleep(time.Duration(domain.RandomInt(10000)) * time.Microsecond)
	}
	return
}

func indexMessage(key, value interface{}) {
	msgIndexer.Enter(key, value)
}

var msgIndexer = domain.NewFlusher(1000, 1, time.Millisecond, func(items []domain.FlusherEntry) {
	_ = domain.Try(100, time.Second, func() error {
		if r.msgSearch == nil {
			return domain.ErrDoesNotExists
		}
		return nil
	})
	b := r.msgSearch.NewBatch()
	for _, item := range items {
		_ = b.Index(item.Key.(string), item.Value)
	}
	err := r.msgSearch.Batch(b)
	if err != nil {
		logs.Warn("MessageIndexer got error", zap.Error(err))
	}
})

func indexMessageRemove(key string) {
	msgIndexRemover.Enter(key, nil)
}

var msgIndexRemover = domain.NewFlusher(1000, 1, time.Millisecond, func(items []domain.FlusherEntry) {
	_ = domain.Try(100, time.Second, func() error {
		if r.msgSearch == nil {
			return domain.ErrDoesNotExists
		}
		return nil
	})
	for _, item := range items {
		_ = r.msgSearch.Delete(item.Key.(string))

	}
})

func indexPeer(key, value interface{}) {
	peerIndexer.Enter(key, value)
}

var peerIndexer = domain.NewFlusher(1000, 1, time.Millisecond, func(items []domain.FlusherEntry) {
	domain.Try(100, time.Second, func() error {
		if r.peerSearch == nil {
			return domain.ErrDoesNotExists
		}
		return nil
	})
	b := r.peerSearch.NewBatch()
	for _, item := range items {
		_ = b.Index(item.Key.(string), item.Value)
	}
	err := r.peerSearch.Batch(b)
	if err != nil {
		logs.Warn("PeerIndexer got error", zap.Error(err))
	}
})

func indexPeerRemove(key string) {
	peerIndexRemover.Enter(key, nil)
}

var peerIndexRemover = domain.NewFlusher(1000, 1, time.Millisecond, func(items []domain.FlusherEntry) {
	domain.Try(100, time.Second, func() error {
		if r.peerSearch == nil {
			return domain.ErrDoesNotExists
		}
		return nil
	})
	for _, item := range items {
		_ = r.peerSearch.Delete(item.Key.(string))

	}
})
