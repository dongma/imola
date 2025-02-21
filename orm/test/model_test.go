package test

import (
	"github.com/stretchr/testify/assert"
	"imola/orm"
	"imola/orm/internal/errs"
	"testing"
)

func TestParseModel(t *testing.T) {
	testCases := []struct {
		name      string
		entity    any
		wantModel *orm.Model
		wantErr   error
	}{
		{
			name:    "struct",
			entity:  TestModel{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "map",
			entity:  map[string]string{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "slice",
			entity:  []int{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:   "pointer",
			entity: &TestModel{},
			wantModel: &orm.Model{
				TableName: "test_model",
				Fields: map[string]*orm.Field{
					"Id": {
						Column: "id",
					},
					"FirstName": {
						Column: "first_name",
					},
					"LastName": {
						Column: "last_name",
					},
					"Age": {
						Column: "age",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := orm.ParseModel(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantModel, m)
		})
	}
}
