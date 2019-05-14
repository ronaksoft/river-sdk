package shared

import (
	"fmt"
	log "git.ronaksoftware.com/ronak/toolbox/logger"
	"sync"
)

var (
	mx                   sync.Mutex
	cachedActorsByAuthID map[int64]Actor
	cachedActorsByPhone  map[string]Actor
)

var (
	_Log log.Logger
)

func init() {
	_Log = log.NewConsoleLogger()
}

func SetLogger(l log.Logger) {
	_Log = l
}

func init() {
	cachedActorsByAuthID = make(map[int64]Actor)
	cachedActorsByPhone = make(map[string]Actor)
}

func CacheActor(act Actor) {
	mx.Lock()
	authID := act.GetAuthID()
	if a, ok := cachedActorsByAuthID[authID]; ok {
		_Log.Warn(fmt.Sprintf("Duplicated AuthID \t Act:%s and Act:%s have equal authID\n\nAct:%x \nAct:%x",
			act.GetPhone(),
			a.GetPhone(),
			act.GetAuthKey(),
			a.GetAuthKey(),
		))
	}
	cachedActorsByAuthID[act.GetAuthID()] = act
	cachedActorsByPhone[act.GetPhone()] = act
	mx.Unlock()
}

func GetCachedActorByAuthID(authID int64) (act Actor, ok bool) {
	mx.Lock()
	act, ok = cachedActorsByAuthID[authID]
	mx.Unlock()
	return
}

func GetCachedActorByPhone(phone string) (act Actor, ok bool) {
	mx.Lock()
	act, ok = cachedActorsByPhone[phone]
	mx.Unlock()
	return
}

func GetCachedAllActors() map[string]Actor {
	return cachedActorsByPhone
}
