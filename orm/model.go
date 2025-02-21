package orm

import (
	"imola/orm/internal/errs"
	"reflect"
	"unicode"
)

type Model struct {
	TableName string
	Fields    map[string]*Field
}

type Field struct {
	// 列名
	Column string
}

// ParseModel 限制只能使用一级指针
func ParseModel(entity any) (*Model, error) {
	typ := reflect.TypeOf(entity)
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}
	typ = typ.Elem()
	numField := typ.NumField()
	fieldMap := make(map[string]*Field, numField)
	for i := 0; i < numField; i++ {
		fd := typ.Field(i)
		fieldMap[fd.Name] = &Field{
			Column: underscoreName(fd.Name),
		}
	}
	return &Model{
		TableName: underscoreName(typ.Name()),
		Fields:    fieldMap,
	}, nil
}

// underscoreName 驼峰转换字符串
func underscoreName(tableName string) string {
	var buf []byte
	for i, v := range tableName {
		if unicode.IsUpper(v) {
			if i != 0 {
				buf = append(buf, '_')
			}
			buf = append(buf, byte(unicode.ToLower(v)))
		} else {
			buf = append(buf, byte(v))
		}
	}
	return string(buf)
}
