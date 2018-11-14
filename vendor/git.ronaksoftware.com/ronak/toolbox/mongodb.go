package ronak

import (
    "crypto/tls"
    "net"
    "sync"
    "time"

    "github.com/globalsign/mgo"
)

// MongoConfig
type MongoConfig struct {
    Host               string
    NumberOfSessions   int
    InsecureSkipVerify bool
    DialTimeout        time.Duration
    DB                 string
}

var (
    DefaultMongoConfig = MongoConfig{
        NumberOfSessions:   10,
        InsecureSkipVerify: false,
        DialTimeout:        10 * time.Second,
    }
)

type MongoDB struct {
    dbs         []*mgo.Database
    lastRefresh []time.Time
}

func NewMongoDB(config MongoConfig) *MongoDB {
    mdb := new(MongoDB)
    mdb.dbs = make([]*mgo.Database, 0, config.NumberOfSessions)
    mdb.lastRefresh = make([]time.Time, 0, config.NumberOfSessions)

    waitGroup := sync.WaitGroup{}
    for i := 0; i < config.NumberOfSessions; i++ {
        waitGroup.Add(1)
        go func(idx int) {
            defer waitGroup.Done()
            // Initialize MongoDB
            tlsConfig := new(tls.Config)
            tlsConfig.InsecureSkipVerify = config.InsecureSkipVerify
            if dialInfo, err := mgo.ParseURL(config.Host); err != nil {
                _LOG.Fatal(err.Error())
            } else {
                dialInfo.Timeout = config.DialTimeout
                dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
                    if conn, err := tls.Dial("tcp", addr.String(), tlsConfig); err != nil {
                        return conn, err
                    } else {
                        return conn, nil
                    }
                }
                if mongoSession, err := mgo.DialWithInfo(dialInfo); err != nil {
                    if mongoSession, err = mgo.Dial(config.Host); err != nil {
                        _LOG.Fatal(err.Error())
                    } else {
                        _LOG.Info("MongoDB Connected")
                        mdb.dbs = append(mdb.dbs, mongoSession.DB(config.DB))
                        mdb.lastRefresh = append(mdb.lastRefresh, time.Now())
                    }
                } else {
                    _LOG.Info("MongoDB(TLS) Connected")
                    mdb.dbs = append(mdb.dbs, mongoSession.DB(config.DB))
                    mdb.lastRefresh = append(mdb.lastRefresh, time.Now())
                }
            }
        }(i)
    }
    waitGroup.Wait()

    if len(mdb.dbs) == 0 {
        _LOG.Fatal("no db session was initiated")
    }

    return mdb
}

func (mdb *MongoDB) GetDB() *mgo.Database {
    idx := RandomInt64(int64(len(mdb.dbs)))
    db := mdb.dbs[idx]

    // Refresh MongoDB Connection at most once per second
    if time.Now().Sub(mdb.lastRefresh[idx]) > time.Second {
        db.Session.Refresh()
        mdb.lastRefresh[idx] = time.Now()
    }
    return db
}
