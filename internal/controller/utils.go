package controller

import (
	"fmt"
	"reflect"
)

func MergeMaps[K comparable, V any](maps ...map[K]V) map[K]V {
	result := make(map[K]V)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// The merge uses extra spaces using a map
func MergeSliceByName[O interface{}](objs ...*[]O) (*[]O, error) {
	mm := make(map[string]O)
	for _, o := range objs {
		if o == nil {
			continue
		}
		for _, mi := range *o {
			// the later one overrides the previous ones
			name, r := GetFieldValue(mi, "Name")
			if r {
				mm[name.(string)] = mi
			} else {
				return nil, fmt.Errorf("there is no Name field in: %v", mi)
			}
		}
	}
	result := make([]O, 0, len(mm))
	for _, v := range mm {
		result = append(result, v)
	}
	return &result, nil
}

func GetFieldValue(obj interface{}, fieldName string) (interface{}, bool) {
	val := reflect.ValueOf(obj)

	// If it's a pointer, get the element it points to
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Must be a struct to have fields
	if val.Kind() != reflect.Struct {
		return nil, false
	}

	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		return nil, false
	}

	return field.Interface(), true
}
