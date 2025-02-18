package reflect

import (
	"errors"
	"reflect"
)

// IterateFields 遍历实体的字段
func IterateFields(entity any) (map[string]any, error) {
	if entity == nil {
		return nil, errors.New("不支持nil")
	}
	typ := reflect.TypeOf(entity)
	val := reflect.ValueOf(entity)
	if val.IsZero() {
		return nil, errors.New("不支持零值")
	}

	// 如果entity是指针类型，拿到指针指向的对象
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
		val = val.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil, errors.New("不支持类型")
	}

	numField := typ.NumField()
	res := make(map[string]any, numField)
	for i := 0; i < numField; i++ {
		// 分别取出字段类型fieldType、字段值fieldVal
		fieldType := typ.Field(i)
		fieldVal := val.Field(i)
		if fieldType.IsExported() {
			res[fieldType.Name] = fieldVal.Interface()
		} else {
			res[fieldType.Name] = reflect.Zero(fieldType.Type).Interface()
		}
	}
	return res, nil
}
