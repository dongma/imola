package sql

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JsonColumn T是一个可被Json处理的类型
type JsonColumn[T any] struct {
	Val T
	// 主要解决Null之类的问题
	Valid bool
}

func (j JsonColumn[T]) Value() (driver.Value, error) {
	// Null
	if !j.Valid {
		return nil, nil
	}
	return json.Marshal(j.Val)
}

func (j *JsonColumn[T]) Scan(src any) error {
	var bs []byte
	switch data := src.(type) {
	case string:
		bs = []byte(data)
	case []byte:
		bs = data
	case nil:
	default:
		return errors.New("不支持的类型")
	}
	err := json.Unmarshal(bs, &j.Val)
	if err == nil {
		j.Valid = true
	}
	return err
}
