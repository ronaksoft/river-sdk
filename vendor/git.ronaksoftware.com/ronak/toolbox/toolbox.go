package ronak

import (
	"git.ronaksoftware.com/ronak/toolbox/logger"
	"reflect"
)

/*
   Creation Time: 2018 - Apr - 07
   Created by:  Ehsan N. Moosa (ehsan)
   Maintainers:
       1.  Ehsan N. Moosa (ehsan)
   Auditor: Ehsan N. Moosa
   Copyright Ronak Software Group 2018
*/

var (
	_Log log.Logger
)

type (
	M map[string]interface{}
	MS map[string]string
	MI map[string]int64
	UniqueArray map[interface{}]struct{}
)

func (m UniqueArray) ToArray() interface{} {
	keys := reflect.ValueOf(m).MapKeys()
	if len(keys) == 0 {
		return reflect.ValueOf(nil)
	}
	x := reflect.MakeSlice(reflect.SliceOf(keys[0].Elem().Type()), 0, len(keys))
	for _, k := range keys {
		x = reflect.Append(x, k.Elem())
	}
	return x.Interface()
}

func init() {
	_Log = log.NewNop()
}

func SetLogger(l log.Logger) {
	_Log = l
}
