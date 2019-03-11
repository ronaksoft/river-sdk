package ronak

// RedisCounterManager
type RedisCounterManager struct {
	redisCache *RedisCache
}

func NewCounterManager(cache *RedisCache) *RedisCounterManager {
	m := new(RedisCounterManager)
	m.redisCache = cache
	return m
}

func (m *RedisCounterManager) Exists(counterName string) bool {
	if b, err := m.redisCache.Exists(counterName); err == nil && b {
		return true
	}
	return false
}

func (m *RedisCounterManager) Inc(counterName string, n int64) (v int64, err error) {
	v, err = m.redisCache.IncBy(counterName, n)
	return
}

func (m *RedisCounterManager) Get(counterName string) (v int64, err error) {
	v, err = m.redisCache.GetInt64(counterName)
	return
}
