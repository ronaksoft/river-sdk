package tools

import (
	"reflect"
)

/*
   Creation Time: 2019 - Oct - 13
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func DeleteItemFromArray(slice interface{}, index int) {
	v := reflect.ValueOf(slice)
	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}
	vLength := v.Len()
	if v.Kind() != reflect.Slice {
		panic("slice is not valid")
	}
	if index >= vLength || index < 0 {
		panic("invalid index")
	}
	switch vLength {
	case 1:
		v.SetLen(0)
	default:
		v.Index(index).Set(v.Index(v.Len() - 1))
		v.SetLen(vLength - 1)
	}
}

func TruncateString(str string, l int) string {
	n := 0
	for i := range str {
		n++
		if n > l {
			return str[:i]
		}
	}
	return str
}
