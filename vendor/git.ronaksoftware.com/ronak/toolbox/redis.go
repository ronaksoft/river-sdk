package ronak

import (
	"fmt"
	"net"
	"time"

	"github.com/mediocregopher/radix/v3"
)

/*
    Creation Time: 2018 - Apr - 07
    Created by:  Ehsan N. Moosa (ehsan)
    Maintainers:
        1.  Ehsan N. Moosa (ehsan)
		2.  Hamidreza KK
    Auditor: Ehsan N. Moosa
    Copyright Ronak Software Group 2018
*/

type connType int

const (
	_ connType = iota
	connTypePool
	connTypeCluster
)

// RedisConfig
type RedisConfig struct {
	ConnectionPoolSize    int
	ConnectionDialTimeout time.Duration
	ConnectionTimeout     time.Duration
	Password              string
	Host                  string
	ClusterHosts          []string
	PingTime              time.Duration
	RefillInterval        time.Duration
	Db                    int
}

type (
	CmdAction radix.CmdAction
	Action radix.Action
)

var (
	DefaultRedisConfig = RedisConfig{
		ConnectionPoolSize:    10,
		ConnectionDialTimeout: 5 * time.Second,
		ConnectionTimeout:     5 * time.Second,
		PingTime:              time.Minute,
		RefillInterval:        time.Second,
		Db:                    0,
	}
)

// RedisCache
type RedisCache struct {
	cluster  *radix.Cluster
	pool     *radix.Pool
	connType connType
	conn     RedisConn
	scripts  map[string]radix.EvalScript
}

type RedisConn interface {
	Do(action radix.Action) error
	Close() error
}

type RedisScanner interface {
	Next(*string) bool
	Close() error
}

// NewRedisCache
// This is the constructor of RedisCache, it accepts RedisConfig as input, you can use
// DefaultRedisConfig for quick initialization, but make sure to add 'Conn' and 'Password' to it
//
// example:
// conf := ronak.DefaultRedisConfig
// conf.Conn = "your-host.com"
// conf.Password = "password123"
// c := NewRedisCache(conf)
func NewRedisCache(conf RedisConfig) *RedisCache {
	r := new(RedisCache)
	r.scripts = make(map[string]radix.EvalScript)
	var poolConnFuncOpt = radix.PoolConnFunc(func(network, addr string) (radix.Conn, error) {
		c, err := net.DialTimeout(network, addr, conf.ConnectionDialTimeout)
		if err != nil {
			return nil, err
		}
		conn := radix.NewConn(c)
		if conf.Db > 0 {
			err = conn.Do(radix.Cmd(nil, "SELECT", fmt.Sprintf("%d", conf.Db)))
			if err != nil {
				c.Close()
				return nil, err
			}
		}

		return conn, nil
	})
	if len(conf.Password) > 0 {
		poolConnFuncOpt = radix.PoolConnFunc(func(network, addr string) (radix.Conn, error) {
			c, err := net.DialTimeout(network, addr, conf.ConnectionDialTimeout)
			if err != nil {
				return nil, err
			}
			conn := radix.NewConn(c)
			err = conn.Do(radix.Cmd(nil, "AUTH", conf.Password))
			if err != nil {
				c.Close()
				return nil, err
			}
			if conf.Db > 0 {
				err = conn.Do(radix.Cmd(nil, "SELECT", fmt.Sprintf("%d", conf.Db)))
				if err != nil {
					c.Close()
					return nil, err
				}
			}
			return conn, nil
		})
	}

	pool, err := radix.NewPool("tcp", conf.Host, conf.ConnectionPoolSize,
		poolConnFuncOpt,
		radix.PoolPingInterval(conf.PingTime/time.Duration(conf.ConnectionPoolSize)),
		radix.PoolOnEmptyErrAfter(conf.ConnectionTimeout),
		radix.PoolRefillInterval(conf.RefillInterval),
	)
	if err != nil {
		_Log.Fatal(err.Error())
	}
	r.conn = pool
	r.connType = connTypePool
	r.pool = pool
	return r
}

func NewRedisCacheWithDB(conf RedisConfig, dbNum int) *RedisCache {
	r := new(RedisCache)
	r.scripts = make(map[string]radix.EvalScript)
	var PoolOpt radix.PoolOpt = nil
	if len(conf.Password) > 0 {
		PoolOpt = radix.PoolConnFunc(func(network, addr string) (radix.Conn, error) {
			c, err := net.Dial(network, addr)
			if err != nil {
				return nil, err
			}
			conn := radix.NewConn(c)
			conn.Do(radix.Cmd(nil, "AUTH", conf.Password))
			err = conn.Do(radix.Cmd(nil, "SELECT", fmt.Sprintf("%d", dbNum)))
			if err != nil {
				c.Close()
				return nil, err
			}
			return conn, nil
		})
	}

	pool, err := radix.NewPool("tcp", conf.Host, 1, PoolOpt)
	if err != nil {
		_Log.Fatal(err.Error())
	}
	r.conn = pool
	r.connType = connTypePool
	r.pool = pool
	return r
}

func NewRedisClusterCache(conf RedisConfig) *RedisCache {
	r := new(RedisCache)
	r.scripts = make(map[string]radix.EvalScript)
	var ClusterOpt radix.ClusterOpt = nil
	if len(conf.Password) > 0 {
		radix.PoolConnFunc(func(network, addr string) (radix.Conn, error) {
			c, err := net.Dial(network, addr)
			if err != nil {
				return nil, err
			}
			conn := radix.NewConn(c)
			conn.Do(radix.Cmd(nil, "AUTH", conf.Password))
			return conn, nil
		})
	}
	cluster, err := radix.NewCluster(conf.ClusterHosts, ClusterOpt)
	if err != nil {
		_Log.Fatal(err.Error())
	}
	r.conn = cluster
	r.connType = connTypeCluster
	r.cluster = cluster
	time.AfterFunc(time.Minute, func() {
		cluster.Sync()
	})

	return r

}

// NewScanner
func (r *RedisCache) NewScanner(opts radix.ScanOpts) RedisScanner {
	switch r.connType {
	case connTypePool:
		return radix.NewScanner(r.conn, opts)
	case connTypeCluster:
		return r.cluster.NewScanner(opts)
	}
	return nil
}

// RegisterScript
func (r *RedisCache) RegisterScript(name string, numKeys int, script string) {
	r.scripts[name] = radix.NewEvalScript(numKeys, script)
}

// RunScript
func (r *RedisCache) RunScript(name string, result *interface{}, args ...string) error {
	return r.Do(r.scripts[name].Cmd(result, args...))
}

// Do
func (r *RedisCache) Do(action radix.Action) error {
	return r.conn.Do(action)
}

func (r *RedisCache) Pipeline(commands ...radix.CmdAction) error {
	return r.Do(radix.Pipeline(commands...))
}

func (r *RedisCache) Cmd(rcv interface{}, cmd string, key string, args ...interface{}) radix.CmdAction {
	return radix.FlatCmd(rcv, cmd, key, args...)
}

func (r *RedisCache) Multi() {
	var rcv string
	r.conn.Do(radix.Cmd(&rcv, "MULTI"))
}

func (r *RedisCache) Exec() {
	var rcv []string
	r.conn.Do(radix.Cmd(&rcv, "EXEC"))
}

func (r *RedisCache) Close() error {
	return r.conn.Close()
}

func (r *RedisCache) Exists(keyName string) (reply bool, err error) {
	err = r.Do(radix.Cmd(&reply, "EXISTS", keyName))
	return
}

func (r *RedisCache) Expire(keyName string, ttl int) (reply bool, err error) {
	err = r.Do(radix.Cmd(&reply, "EXPIRE", keyName, fmt.Sprintf("%d", ttl)))
	return
}

func (r *RedisCache) Del(keyName ...string) (reply bool, err error) {
	err = r.Do(radix.Cmd(&reply, "DEL", keyName...))
	return
}

func (r *RedisCache) Set(keyName string, value interface{}) (reply bool, err error) {
	err = r.Do(radix.FlatCmd(&reply, "SET", keyName, value))
	return
}

func (r *RedisCache) SetNx(keyName string, value interface{}) (reply bool, err error) {
	err = r.Do(radix.FlatCmd(&reply, "SETNX", keyName, value))
	return
}

func (r *RedisCache) SetEx(keyName string, ttl, value interface{}) (reply bool, err error) {
	err = r.Do(radix.FlatCmd(&reply, "SETEX", keyName, ttl, value))
	return
}

func (r *RedisCache) GetString(keyName string) (reply string, err error) {
	err = r.Do(radix.Cmd(&reply, "GET", keyName))
	return
}

func (r *RedisCache) GetInt(keyName string) (reply int, err error) {
	err = r.Do(radix.Cmd(&reply, "GET", keyName))
	return
}

func (r *RedisCache) GetInt32(keyName string) (reply int32, err error) {
	err = r.Do(radix.Cmd(&reply, "GET", keyName))
	return
}

func (r *RedisCache) GetInt64(keyName string) (reply int64, err error) {
	err = r.Do(radix.Cmd(&reply, "GET", keyName))
	return
}

func (r *RedisCache) GetUInt64(keyName string) (reply uint64, err error) {
	err = r.Do(radix.Cmd(&reply, "GET", keyName))
	return
}

func (r *RedisCache) GetBytes(keyName string) (reply []byte, err error) {
	err = r.Do(radix.Cmd(&reply, "GET", keyName))
	return
}

func (r *RedisCache) GetByteSlice(keyName string) (reply [][]byte, err error) {
	err = r.Do(radix.Cmd(&reply, "GET", keyName))
	return
}

func (r *RedisCache) MGetBytes(keyNames ...string) (reply [][]byte, err error) {
	err = r.Do(radix.Cmd(&reply, "MGET", keyNames...))
	return
}

func (r *RedisCache) Inc(keyName string) (reply interface{}, err error) {
	err = r.Do(radix.Cmd(&reply, "INCR", keyName))
	return
}

func (r *RedisCache) IncInt64(keyName string) (reply int64, err error) {
	err = r.Do(radix.Cmd(&reply, "INCR", keyName))
	return
}

func (r *RedisCache) IncBy(keyName string, n int64) (reply int64, err error) {
	err = r.Do(radix.Cmd(&reply, "INCRBY", keyName, fmt.Sprintf("%d", n)))
	return
}

func (r *RedisCache) HSet(keyName string, fieldName interface{}, value interface{}) (reply bool, err error) {
	err = r.Do(radix.FlatCmd(&reply, "HSET", keyName, fieldName, value))
	return
}

func (r *RedisCache) HMSetStringMap(hashName string, kv MS) (reply string, err error) {
	args := make([]string, 0, len(kv)*2+1)
	args = append(args, hashName)
	for key, value := range kv {
		args = append(args, key, value)
	}
	err = r.Do(radix.Cmd(&reply, "HMSET", args...))
	return
}

func (r *RedisCache) HMSetInt64Map(hashName string, kv MI) (reply string, err error) {
	args := make([]string, 0, len(kv)*2+1)
	args = append(args, hashName)
	for key, value := range kv {
		args = append(args, key, fmt.Sprintf("%d", value))
	}
	err = r.Do(radix.Cmd(&reply, "HMSET", args...))
	return
}

func (r *RedisCache) HExists(keyName string, fieldName string) (reply bool, err error) {
	err = r.Do(radix.Cmd(&reply, "HEXISTS", keyName, fieldName))
	return
}

func (r *RedisCache) HGetStrings(keyName string, fieldName string) (reply []string, err error) {
	err = r.Do(radix.Cmd(&reply, "HGET", keyName, fieldName))
	return
}

func (r *RedisCache) HGetBytes(keyName string, fieldName string) (reply []byte, err error) {
	err = r.Do(radix.Cmd(&reply, "HGET", keyName, fieldName))
	return
}

func (r *RedisCache) HGetString(keyName string, fieldName string) (reply string, err error) {
	err = r.Do(radix.Cmd(&reply, "HGET", keyName, fieldName))
	return
}

func (r *RedisCache) HGetInt(keyName string, fieldName string) (reply int, err error) {
	err = r.Do(radix.Cmd(&reply, "HGET", keyName, fieldName))
	return
}

func (r *RedisCache) HGetInt32(keyName string, fieldName string) (reply int32, err error) {
	err = r.Do(radix.Cmd(&reply, "HGET", keyName, fieldName))
	return
}

func (r *RedisCache) HGetInt64(keyName string, fieldName string) (reply int64, err error) {
	err = r.Do(radix.Cmd(&reply, "HGET", keyName, fieldName))
	return
}

func (r *RedisCache) HGetUInt64(keyName string, fieldName string) (reply uint64, err error) {
	err = r.Do(radix.Cmd(&reply, "HGET", keyName, fieldName))
	return
}

func (r *RedisCache) HGetAllStringMap(keyName string) (reply map[string]string, err error) {
	err = r.Do(radix.Cmd(&reply, "HGETALL", keyName))
	return
}

func (r *RedisCache) HGetAllInt64Map(keyName string) (reply map[string]int64, err error) {
	err = r.Do(radix.Cmd(&reply, "HGETALL", keyName))
	return
}

func (r *RedisCache) HGetAllInt32Map(keyName string) (reply map[string]int32, err error) {
	err = r.Do(radix.Cmd(&reply, "HGETALL", keyName))
	return
}

func (r *RedisCache) HGetAll(keyName string) (reply map[string]interface{}, err error) {
	err = r.Do(radix.Cmd(&reply, "HGETALL", keyName))
	return
}

func (r *RedisCache) HIncrementBy(keyName string, fieldName string, incr int) (reply int, err error) {
	err = r.Do(radix.Cmd(&reply, "HINCRBY", keyName, fieldName, fmt.Sprintf("%d", incr)))
	return
}

func (r *RedisCache) HKeysStrings(keyName string) (reply []string, err error) {
	err = r.Do(radix.Cmd(&reply, "HKEYS", keyName))
	return
}

func (r *RedisCache) HKeysInts(keyName string) (reply []int, err error) {
	err = r.Do(radix.Cmd(&reply, "HKEYS", keyName))
	return
}

func (r *RedisCache) HKeysInt32s(keyName string) (reply []int32, err error) {
	err = r.Do(radix.Cmd(&reply, "HKEYS", keyName))
	return
}

func (r *RedisCache) HKeysInt64s(keyName string) (reply []int64, err error) {
	err = r.Do(radix.Cmd(&reply, "HKEYS", keyName))
	return
}

func (r *RedisCache) HValuesStrings(keyName string) (reply []string, err error) {
	err = r.Do(radix.Cmd(&reply, "HVALS", keyName))
	return
}

func (r *RedisCache) HValuesInts(keyName string) (reply []int, err error) {
	err = r.Do(radix.Cmd(&reply, "HVALS", keyName))
	return
}

func (r *RedisCache) HValuesInt32s(keyName string) (reply []int32, err error) {
	err = r.Do(radix.Cmd(&reply, "HVALS", keyName))
	return
}

func (r *RedisCache) HValuesInt64s(keyName string) (reply []int64, err error) {
	err = r.Do(radix.Cmd(&reply, "HVALS", keyName))
	return
}

func (r *RedisCache) HDel(keyName string, fieldName string) (reply bool, err error) {
	err = r.Do(radix.Cmd(&reply, "HDEL", keyName, fieldName))
	return
}

func (r *RedisCache) RPushBytes(keyName string, v []byte) (reply int, err error) {
	err = r.Do(radix.Cmd(&reply, "RPUSH", keyName, string(v)))
	return
}

func (r *RedisCache) RPushInt(keyName string, v int) (reply int, err error) {
	err = r.Do(radix.Cmd(&reply, "RPUSH", keyName, fmt.Sprintf("%d", v)))
	return
}

func (r *RedisCache) RPushInt32(keyName string, v int32) (reply int, err error) {
	err = r.Do(radix.Cmd(&reply, "RPUSH", keyName, fmt.Sprintf("%d", v)))
	return
}

func (r *RedisCache) RPushInt64(keyName string, v int64) (reply int, err error) {
	err = r.Do(radix.Cmd(&reply, "RPUSH", keyName, fmt.Sprintf("%d", v)))
	return
}

func (r *RedisCache) RPopBytes(keyName string) (reply []byte, err error) {
	err = r.Do(radix.Cmd(&reply, "RPOP", keyName))
	return
}

func (r *RedisCache) BRPopBytes(keyName string, time int) (reply []byte, err error) {
	err = r.Do(radix.Cmd(&reply, "BRPOP", keyName, fmt.Sprintf("%d", time)))
	return
}

func (r *RedisCache) LPushBytes(keyName string, v []byte) (reply int, err error) {
	err = r.Do(radix.Cmd(&reply, "LPUSH", keyName, string(v)))
	return
}

func (r *RedisCache) LPushInt(keyName string, v int) (reply int, err error) {
	err = r.Do(radix.Cmd(&reply, "LPUSH", keyName, fmt.Sprintf("%d", v)))
	return
}

func (r *RedisCache) LPushInt32(keyName string, v int32) (reply int, err error) {
	err = r.Do(radix.Cmd(&reply, "LPUSH", keyName, fmt.Sprintf("%d", v)))
	return
}

func (r *RedisCache) LPushInt64(keyName string, v int64) (reply int, err error) {
	err = r.Do(radix.Cmd(&reply, "LPUSH", keyName, fmt.Sprintf("%d", v)))
	return
}

func (r *RedisCache) LPopBytes(keyName string) (reply []byte, err error) {
	err = r.Do(radix.Cmd(&reply, "LPOP", keyName))
	return
}

func (r *RedisCache) BLPop(keyName string, time int) (reply []byte, err error) {
	err = r.Do(radix.Cmd(&reply, "BLPOP", keyName, fmt.Sprintf("%d", time)))
	return
}

func (r *RedisCache) LLen(keyName string) (reply int, err error) {
	err = r.Do(radix.Cmd(&reply, "LLEN", keyName))
	return
}

func (r *RedisCache) SAdd(keyName string, values ...interface{}) (reply int, err error) {
	err = r.Do(radix.FlatCmd(&reply, "SADD", keyName, values...))
	return
}

func (r *RedisCache) SAddString(keyName string, value string) (reply int, err error) {
	err = r.Do(radix.Cmd(&reply, "SADD", keyName, value))
	return
}

func (r *RedisCache) SAddBytes(keyName string, value []byte) (reply int, err error) {
	err = r.Do(radix.FlatCmd(&reply, "SADD", keyName, value))
	return
}

func (r *RedisCache) SAddInt(keyName string, value int) (reply int, err error) {
	err = r.Do(radix.FlatCmd(&reply, "SADD", keyName, value))
	return
}

func (r *RedisCache) SAddInt32(keyName string, value int32) (reply int, err error) {
	err = r.Do(radix.FlatCmd(&reply, "SADD", keyName, value))
	return
}

func (r *RedisCache) SAddInt64(keyName string, value int64) (reply int, err error) {
	err = r.Do(radix.FlatCmd(&reply, "SADD", keyName, value))
	return
}

func (r *RedisCache) SAddUInt64(keyName string, values ...uint64) (reply int, err error) {
	err = r.Do(radix.FlatCmd(&reply, "SADD", keyName, values))
	return
}

func (r *RedisCache) SRemove(keyName string, value interface{}) (reply int, err error) {
	err = r.Do(radix.FlatCmd(&reply, "SREM", keyName, value))
	return
}

func (r *RedisCache) SRemString(keyName string, value string) (reply int, err error) {
	err = r.Do(radix.Cmd(&reply, "SREM", keyName, value))
	return
}

func (r *RedisCache) SRemBytes(keyName string, value []byte) (reply int, err error) {
	err = r.Do(radix.Cmd(&reply, "SREM", keyName, string(value)))
	return
}

func (r *RedisCache) SRemInt(keyName string, value int) (reply int, err error) {
	err = r.Do(radix.Cmd(&reply, "SREM", keyName, fmt.Sprintf("%d", value)))
	return
}

func (r *RedisCache) SRemInt32(keyName string, value int32) (reply int, err error) {
	err = r.Do(radix.Cmd(&reply, "SREM", keyName, fmt.Sprintf("%d", value)))
	return
}

func (r *RedisCache) SRemInt64(keyName string, value int64) (reply int, err error) {
	err = r.Do(radix.Cmd(&reply, "SREM", keyName, fmt.Sprintf("%d", value)))
	return
}

func (r *RedisCache) SScan(keyName string, cursorID int) {

}

func (r *RedisCache) SMembersInt64s(keyName string) (reply []int64, err error) {
	err = r.Do(radix.Cmd(&reply, "SMEMBERS", keyName))
	return
}

func (r *RedisCache) SMembersUInt64s(keyName string) (reply []uint64, err error) {
	err = r.Do(radix.Cmd(&reply, "SMEMBERS", keyName))
	return
}

func (r *RedisCache) SMembersStrings(keyName string) (reply []string, err error) {
	err = r.Do(radix.Cmd(&reply, "SMEMBERS", keyName))
	return
}

func (r *RedisCache) SIsMember(keyName string, value interface{}) (reply bool, err error) {
	err = r.Do(radix.FlatCmd(&reply, "SISMEMBER", keyName, value))
	return
}

func (r *RedisCache) SCard(keyName string) (reply int, err error) {
	err = r.Do(radix.Cmd(&reply, "SCARD", keyName))
	return
}

func (r RedisCache) SDiffStore(dstKeyName, srcKeyName string, keyNames ...string) (reply int, err error) {
	args := []string{dstKeyName, srcKeyName}
	args = append(args, keyNames...)
	err = r.Do(radix.Cmd(&reply, "SDIFFSTORE", args...))
	return
}

func (r *RedisCache) SDiff(srcKeyName string, keyNames ...string) (reply []string, err error) {
	args := []string{srcKeyName}
	args = append(args, keyNames...)
	err = r.Do(radix.Cmd(&reply, "SDIFF", args...))
	return
}

func (r *RedisCache) SDiffUInt64s(srcKeyName string, keyNames ...string) (reply []uint64, err error) {
	args := []string{srcKeyName}
	args = append(args, keyNames...)
	err = r.Do(radix.Cmd(&reply, "SDIFF", args...))
	return
}

func (r *RedisCache) SDiffInt64s(srcKeyName string, keyNames ...string) (reply []int64, err error) {
	args := []string{srcKeyName}
	args = append(args, keyNames...)
	err = r.Do(radix.Cmd(&reply, "SDIFF", args...))
	return
}

func (r *RedisCache) ZAddString(keyName string, score int, value string) (reply int, err error) {
	err = r.Do(radix.Cmd(&reply, "ZADD", keyName, fmt.Sprintf("%d", score), value))
	return
}

func (r *RedisCache) ZAdd(keyName string, score, value interface{}) (reply int, err error) {
	err = r.Do(radix.FlatCmd(&reply, "ZADD", keyName, score, value))
	return
}

func (r *RedisCache) ZRem(keyName string, field interface{}) (reply int, err error) {
	err = r.Do(radix.FlatCmd(&reply, "ZREM", keyName, field))
	return
}

func (r *RedisCache) ZCard(keyName string) (reply int, err error) {
	err = r.Do(radix.Cmd(&reply, "ZCARD", keyName))
	return
}

func (r *RedisCache) ZRangeStrings(keyName string, start, stop int) (reply []string, err error) {
	err = r.Do(radix.Cmd(&reply, "ZRANGE", keyName, fmt.Sprintf("%d", start), fmt.Sprintf("%d", stop)))
	return
}
