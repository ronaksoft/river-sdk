package shared

import (
	"sync"
)

var (
	mx                   sync.Mutex
	cachedActorsByAuthID map[int64]Acter
	cachedActorsByPhone  map[string]Acter
)

func init() {
	cachedActorsByAuthID = make(map[int64]Acter)
	cachedActorsByPhone = make(map[string]Acter)
}

func CacheActor(act Acter) {
	mx.Lock()
	cachedActorsByAuthID[act.GetAuthID()] = act
	cachedActorsByPhone[act.GetPhone()] = act
	mx.Unlock()
}

func GetCacheActorByAuthID(authID int64) (act Acter, ok bool) {
	mx.Lock()
	act, ok = cachedActorsByAuthID[authID]
	mx.Unlock()
	return
}

func GetCacheActorByPhone(phone string) (act Acter, ok bool) {
	mx.Lock()
	act, ok = cachedActorsByPhone[phone]
	mx.Unlock()
	return
}
