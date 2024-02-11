package repo

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "time"

    "github.com/blevesearch/bleve/v2"
    "github.com/blevesearch/bleve/v2/analysis/analyzer/keyword"
    "github.com/blevesearch/bleve/v2/analysis/lang/en"
    "github.com/blevesearch/bleve/v2/mapping"
    "github.com/dgraph-io/badger/v2"
    "github.com/dgraph-io/badger/v2/options"
    "github.com/pkg/errors"
    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/river-sdk/internal/logs"
    "github.com/ronaksoft/rony/tools"
    "github.com/tidwall/buntdb"
    "go.uber.org/zap"
)

var (
    ctx       *Context
    r         *repository
    singleton sync.Mutex
    logger    *logs.Logger

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
    Notifications   *repoNotifications
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

func MustInit(dbPath string, lowMemory bool) {
    bleve.NewIndexMapping()
    err := Init(dbPath, lowMemory)
    if err != nil {
        panic(err)
    }
}

// Init initialize repo singleton
func Init(dbPath string, lowMemory bool) error {
    if ctx == nil {
        singleton.Lock()
        err := repoSetDB(dbPath, lowMemory)
        if err != nil {
            return err
        }

        logger = logs.With("REPO")
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
        Notifications = &repoNotifications{repository: r}
        singleton.Unlock()
    }
    return nil
}

func repoSetDB(dbPath string, lowMemory bool) error {
    r = new(repository)

    _ = os.MkdirAll(dbPath, os.ModePerm)
    // Initialize BadgerDB
    badgerPath := filepath.Join(dbPath, "badger")
    _ = os.MkdirAll(badgerPath, os.ModePerm)
    badgerOpts := badger.DefaultOptions(badgerPath).
        WithLogger(nil).
        WithTruncate(true)
    if lowMemory {
        badgerOpts = badgerOpts.
            WithTableLoadingMode(options.FileIO).
            WithValueLogLoadingMode(options.FileIO).
            WithNumMemtables(2).
            WithNumLevelZeroTables(2).
            WithNumLevelZeroTablesStall(4).
            WithMaxTableSize(1 << 22). // 4MB
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
    buntPath := filepath.Join(dbPath, "bunty")
    _ = os.MkdirAll(buntPath, os.ModePerm)
    if buntIndex, err := buntdb.Open(fmt.Sprintf("%s/bunty/dialogs.db", strings.TrimRight(dbPath, "/"))); err != nil {
        return err
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
        _ = tools.Try(10, time.Millisecond*100, func() error {
            searchDbPath := fmt.Sprintf("%s/searchdb/msg", strings.TrimRight(dbPath, "/"))
            if msgSearch, err := bleve.Open(searchDbPath); err != nil {
                switch err {
                case bleve.ErrorIndexPathDoesNotExist:
                    // create a mapping
                    r.msgSearch, err = bleve.New(searchDbPath, indexMapForMessages())
                    if err != nil {
                        _ = os.RemoveAll(searchDbPath)
                        return err
                    }
                default:
                    _ = os.RemoveAll(searchDbPath)
                    return err
                }
            } else {
                r.msgSearch = msgSearch
            }
            return nil
        })
    }()
    go func() {
        // 2. Peer Search
        _ = tools.Try(10, 100*time.Millisecond, func() error {
            peerDbPath := fmt.Sprintf("%s/searchdb/peer", strings.TrimRight(dbPath, "/"))
            if peerSearch, err := bleve.Open(peerDbPath); err != nil {
                switch err {
                case bleve.ErrorIndexPathDoesNotExist:
                    // create a mapping
                    r.peerSearch, err = bleve.New(peerDbPath, indexMapForPeers())
                    if err != nil {
                        _ = os.RemoveAll(peerDbPath)
                        return err
                    }
                default:
                    _ = os.RemoveAll(peerDbPath)
                    return err
                }
            } else {
                r.peerSearch = peerSearch
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
        logger.Info("Badger ValueLog GC executed")
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
        default:
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

type keyValue struct {
    Key   interface{}
    Value interface{}
}

func indexMessage(key, value interface{}) {
    msgIndexer.Enter("", tools.NewEntry(&keyValue{
        Key:   key,
        Value: value,
    }))
}

var msgIndexer = tools.NewFlusherPool(10, 1000, func(targetID string, entries []tools.FlushEntry) {
    _ = tools.Try(100, time.Second, func() error {
        if r.msgSearch == nil {
            return domain.ErrDoesNotExists
        }
        return nil
    })
    b := r.msgSearch.NewBatch()
    for _, item := range entries {
        kv := item.Value().(*keyValue)
        _ = b.Index(kv.Key.(string), kv.Value)
    }
    err := r.msgSearch.Batch(b)
    if err != nil {
        logger.Warn("got error MessageIndexer", zap.Error(err))
    }
})

func indexMessageRemove(key string) {
    msgIndexRemover.Enter("", tools.NewEntry(key))
}

var msgIndexRemover = tools.NewFlusherPool(10, 1000, func(targetID string, entries []tools.FlushEntry) {
    _ = tools.Try(100, time.Second, func() error {
        if r.msgSearch == nil {
            return domain.ErrDoesNotExists
        }
        return nil
    })
    for _, item := range entries {
        _ = r.msgSearch.Delete(item.Value().(string))

    }
})

func indexPeer(key, value interface{}) {
    peerIndexer.Enter("", tools.NewEntry(&keyValue{
        Key:   key,
        Value: value,
    }))
}

var peerIndexer = tools.NewFlusherPool(10, 1000, func(targetID string, entries []tools.FlushEntry) {
    _ = tools.Try(100, time.Second, func() error {
        if r.peerSearch == nil {
            return domain.ErrDoesNotExists
        }
        return nil
    })
    b := r.peerSearch.NewBatch()
    for _, item := range entries {
        kv := item.Value().(*keyValue)
        _ = b.Index(kv.Key.(string), kv.Value)
    }
    err := r.peerSearch.Batch(b)
    if err != nil {
        logger.Warn("PeerIndexer got error", zap.Error(err))
    }
})
