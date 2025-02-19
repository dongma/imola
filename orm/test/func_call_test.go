package test

import (
	"github.com/stretchr/testify/assert"
	reflect2 "imola/orm/reflect"
	"imola/orm/reflect/types"
	"reflect"
	"testing"
)

func TestIterateFunc(t *testing.T) {
	testCases := []struct {
		name    string
		entity  any
		wantRes map[string]reflect2.FuncInfo
		wantErr error
	}{
		{
			name:   "struct",
			entity: types.NewUser("Tom", 18),
			wantRes: map[string]reflect2.FuncInfo{
				"GetAge": {
					Name: "GetAge",
					//InputTypes: nil,
					OutputTypes: []reflect.Type{reflect.TypeOf(0)},
					Result:      []any{18},
				},
				"ChangeName": {
					Name:       "ChangeName",
					InputTypes: []reflect.Type{reflect.TypeOf("")},
					//OutputTypes: []reflect.Type{reflect.TypeOf(0)},
					//Result:      []any{18},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := reflect2.IterateFunc(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}

}
