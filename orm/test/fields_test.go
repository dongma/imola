package test

import (
	"errors"
	reflect2 "github.com/dongma/imola/orm/reflect"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestIterateFields(t *testing.T) {
	type User struct {
		Name string
		age  int
	}
	testCases := []struct {
		name    string
		entity  any
		wantErr error
		wantRes map[string]any
	}{
		{
			name: "struct",
			entity: User{
				Name: "Tom",
				age:  18,
			},
			wantRes: map[string]any{
				"Name": "Tom",
				// age是私有的，拿不到，最终我们会创建0值来填充
				"age": 0,
			},
		},
		{
			name: "pointer",
			entity: &User{
				Name: "Tom",
				age:  18,
			},
			wantRes: map[string]any{
				"Name": "Tom",
				// age是私有的，拿不到，最终我们会创建0值来填充
				"age": 0,
			},
		},
		{
			name:    "basic types",
			entity:  18,
			wantErr: errors.New("不支持类型"),
		},
		{
			name: "multiple pointer",
			entity: func() **User {
				res := &User{
					Name: "Tom",
					age:  18,
				}
				return &res
			}(),
			wantRes: map[string]any{
				"Name": "Tom",
				// age是私有的，拿不到，最终我们会创建0值来填充
				"age": 0,
			},
		},
		{
			name:    "nil",
			entity:  nil,
			wantErr: errors.New("不支持nil"),
		},
		{
			name:    "user nil",
			entity:  (*User)(nil),
			wantErr: errors.New("不支持零值"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := reflect2.IterateFields(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}

}

func TestSetField(t *testing.T) {
	type User struct {
		Name string
		age  int
	}

	testCases := []struct {
		name    string
		entity  any
		field   string
		newVal  any
		wantErr error
		// 修改后的entity
		wantEntity any
	}{
		{
			name: "struct",
			entity: User{
				Name: "Tom",
			},
			field:   "Name",
			newVal:  "Jerry",
			wantErr: errors.New("不可修改字段"),
		},
		{
			name: "pointer",
			entity: &User{
				Name: "Tom",
			},
			field:  "Name",
			newVal: "Jerry",
			wantEntity: &User{
				Name: "Jerry",
			},
		},
		{
			name: "pointer exported",
			entity: &User{
				age: 18,
			},
			field:   "age",
			newVal:  19,
			wantErr: errors.New("不可修改字段"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := reflect2.SetField(tc.entity, tc.field, tc.newVal)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantEntity, tc.entity)
		})
	}

	var i = 0
	ptr := &i
	reflect.ValueOf(ptr).Elem().Set(reflect.ValueOf(12))
	assert.Equal(t, 12, i)
}
