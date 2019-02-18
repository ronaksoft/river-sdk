package shared

import (
	"fmt"
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
	authID := act.GetAuthID()
	if a, ok := cachedActorsByAuthID[authID]; ok {
		panic(fmt.Sprintf("Duplicated AuthID \t Act:%s and Act:%s have equal authID\n\nAct:%x \nAct:%x",
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

func GetCachedActorByAuthID(authID int64) (act Acter, ok bool) {
	mx.Lock()
	act, ok = cachedActorsByAuthID[authID]
	mx.Unlock()
	return
}

func GetCachedActorByPhone(phone string) (act Acter, ok bool) {
	mx.Lock()
	act, ok = cachedActorsByPhone[phone]
	mx.Unlock()
	return
}

func GetCachedAllActors() map[string]Acter {
	return cachedActorsByPhone
}
