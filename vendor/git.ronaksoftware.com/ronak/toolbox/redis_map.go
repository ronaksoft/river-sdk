package ronak

import (
    "fmt"
    "github.com/mediocregopher/radix.v3"
)

/*
    Creation Time: 2018 - Apr - 19
    Created by:  (ehsan)
    Maintainers:
        1.  (ehsan)
    Auditor: Ehsan N. Moosa
    Copyright Ronak Software Group 2018
*/

type RedisMapManager struct {
    redisCache *RedisCache
}

func NewRedisMapManager(redisCache *RedisCache) *RedisMapManager {
    m := new(RedisMapManager)
    m.redisCache = redisCache
    return m
}

func (m *RedisMapManager) Exists(mapName string) bool {
    if b, err := m.redisCache.Exists(mapName); err == nil && b {
        return true
    }
    return false
}

func (m *RedisMapManager) SetString(mapName string, key string, value string) (new bool, err error) {
    new, err = m.redisCache.HSetString(mapName, key, value)
    return
}

func (m *RedisMapManager) SetBytes(mapName string, key string, value []byte) (new bool, err error) {
    new, err = m.redisCache.HSetString(mapName, key, string(value))
    return
}

func (m *RedisMapManager) SetInt(mapName string, key string, value int) (new bool, err error) {
    new, err = m.redisCache.HSetString(mapName, key, fmt.Sprintf("%d", value))
    return
}

func (m *RedisMapManager) SetMultiString(mapName string, kv MS) error {
    _, err := m.redisCache.HMSetStringMap(mapName, kv)
    return err
}

func (m *RedisMapManager) SetMultiInt64(mapName string, kv MI) error {
    _, err := m.redisCache.HMSetInt64Map(mapName, kv)
    return err
}

func (m *RedisMapManager) GetBytes(mapName string, key string) (reply []byte, err error) {
    err = m.redisCache.Do(radix.Cmd(&reply, "HGET", mapName, key))
    return
}

func (m *RedisMapManager) GetInt(mapName string, key string) (int, error) {
    return m.redisCache.HGetInt(mapName, key)
}

func (m *RedisMapManager) GetInt64(mapName string, key string) (int64, error) {
    return m.redisCache.HGetInt64(mapName, key)
}

func (m *RedisMapManager) GetString(mapName string, key string) (string, error) {
    return m.redisCache.HGetString(mapName, key)
}

func (m *RedisMapManager) GetMapInt64(mapName string) (reply map[string]int64, err error) {
    return m.redisCache.HGetAllInt64Map(mapName)
}

func (m *RedisMapManager) GetMapString(mapName string) (map[string]string, error) {
    return m.redisCache.HGetAllStringMap(mapName)
}

func (m *RedisMapManager) Inc(mapName string, key string, n int) (reply int, err error) {
    return m.redisCache.HIncrementBy(mapName, key, n)
}

func (m *RedisMapManager) Delete(mapName string) error {
    _, err := m.redisCache.Del(mapName)
    return err

}

func (m *RedisMapManager) DeleteKey(mapName string, key string) error {
    _, err := m.redisCache.HDel(mapName, key)
    return err

}
