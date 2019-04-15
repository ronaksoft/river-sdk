package ronak

import (
	"github.com/mediocregopher/radix/v3"
)

var (
	luaPopAll = radix.NewEvalScript(
		1,
		`
        local result = {}
        local length = tonumber(redis.call('LLEN', KEYS[1]))
        for i = 1 , length do
            local val = redis.call('LPOP',KEYS[1])
            if val then
                table.insert(result,val)
            end
        end
        return result
`)
	luaPushWithLimit = radix.NewEvalScript(
		2,
		`
        local result = 0
        local length = tonumber(redis.call('LLEN', KEYS[1]))
        if length < KEYS[2]  then
        result = redis.call('LPUSH', KEYS[1], KEYS[2])
        end
        return result
`)
)

type RedisQueueManager struct {
	redisCache *RedisCache
}

func NewRedisQueueManager(redisCache *RedisCache) *RedisQueueManager {
	m := new(RedisQueueManager)
	m.redisCache = redisCache

	return m
}

func (m *RedisQueueManager) Exists(queueName string) bool {
	if b, err := m.redisCache.Exists(queueName); err == nil && b {
		return true
	}
	return false
}

func (m *RedisQueueManager) PushBytes(queueName string, item []byte) (size int, err error) {
	return m.redisCache.LPushBytes(queueName, item)
}

func (m *RedisQueueManager) PushWithLimit(queueName string, item string, limit int) (size int, err error) {
	err = m.redisCache.Do(luaPushWithLimit.Cmd(&size, queueName, item))
	return
}

func (m *RedisQueueManager) Pop(queueName string) (b []byte, err error) {
	return m.redisCache.RPopBytes(queueName)
}

func (m *RedisQueueManager) PopAll(queueName string) (b [][]byte, err error) {
	err = m.redisCache.Do(luaPopAll.Cmd(&b, queueName))
	return
}

func (m *RedisQueueManager) Length(queueName string) (l int, err error) {
	return m.redisCache.LLen(queueName)
}
