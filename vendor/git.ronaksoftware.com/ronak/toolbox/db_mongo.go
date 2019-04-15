package ronak

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/globalsign/mgo"
)

type BulkError interface {
	Error() string
	Cases() []mgo.BulkErrorCase
}

// MongoConfig
type MongoConfig struct {
	Host               string
	SessionPoolSize    int
	InsecureSkipVerify bool
	Secure             bool
	DialTimeout        time.Duration
	DB                 string
}

type MongoDB struct {
	s *mgo.Session
}

var (
	DefaultMongoConfig = MongoConfig{
		InsecureSkipVerify: false,
		DialTimeout:        10 * time.Second,
		SessionPoolSize:    4096,
	}
)

func NewMongoDB(config MongoConfig) (*MongoDB, error) {
	mdb := new(MongoDB)
	// Initialize MongoDB
	tlsConfig := new(tls.Config)
	tlsConfig.InsecureSkipVerify = config.InsecureSkipVerify
	dialInfo, err := mgo.ParseURL(config.Host)
	if err != nil {
		_Log.Fatal(err.Error())
	}
	dialInfo.Timeout = config.DialTimeout
	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		if conn, err := tls.Dial("tcp", addr.String(), tlsConfig); err != nil {
			return conn, err
		} else {
			return conn, nil
		}
	}
	mongoSession, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		if mongoSession, err = mgo.Dial(config.Host); err != nil {
			return nil, err
		} else {
			_Log.Info("MongoDB Connected")
			mdb.s = mongoSession
			return mdb, nil
		}
	}
	_Log.Info("MongoDB(TLS) Connected")

	mdb.s = mongoSession
	return mdb, nil
}

func (mdb *MongoDB) GetSession() *mgo.Session {
	return mdb.s
}

func (mdb *MongoDB) Clone() *mgo.Session {
	return mdb.s.Clone()
}

func (mdb *MongoDB) Copy() *mgo.Session {
	return mdb.s.Copy()
}
