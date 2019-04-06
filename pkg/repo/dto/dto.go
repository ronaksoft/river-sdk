package dto

import (
	"encoding/json"
	"reflect"
)

// base struct
type dto struct {
}

func (d *dto) Map(to interface{}) {
	Map(d, to)
}

// Map is up to 5x faster than MapJson
func Map(from interface{}, to interface{}) {
	srcElement := reflect.ValueOf(from).Elem()
	desElement := reflect.ValueOf(to).Elem()
	srcCount := srcElement.NumField()
	desCount := srcElement.NumField()
	max := srcCount
	if desCount > srcCount {
		max = desCount
		tmp := desElement
		srcElement = desElement
		desElement = tmp
	}
	srcTyp := srcElement.Type()
	desTyp := desElement.Type()

	for i := 0; i < max; i++ {
		sVal := srcElement.Field(i)
		sTyp := srcTyp.Field(i)

		dTyp, ok := desTyp.FieldByName(sTyp.Name)
		if !ok || dTyp.Type != sTyp.Type {
			continue
		}
		// FieldByIndex is way more faster than FieldByName
		dVal := desElement.FieldByIndex(dTyp.Index)
		if dVal.CanSet() {
			dVal.Set(sVal)
		}

	}
}

// MapJSON uses json tag to map
func MapJSON(from interface{}, to interface{}) error {
	buff, err := json.Marshal(from)
	if err != nil {
		return err
	}
	err = json.Unmarshal(buff, to)
	return err
}
