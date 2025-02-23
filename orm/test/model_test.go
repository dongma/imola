package test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"imola/orm"
	"imola/orm/internal/errs"
	"reflect"
	"testing"
)

func TestParseModel(t *testing.T) {
	testCases := []struct {
		name      string
		entity    any
		wantModel *orm.Model
		wantErr   error
		opts      []orm.ModelOpt
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

	register := orm.Registry{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := register.Register(tc.entity, tc.opts...)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantModel, m)
		})
	}
}

func TestRegistry_get(t *testing.T) {
	testCases := []struct {
		name      string
		entity    any
		wantModel *orm.Model
		wantErr   error
	}{
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
		{
			name: "tag",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column=first_name_t"`
				}
				return &TagTable{}
			}(),
			wantModel: &orm.Model{
				TableName: "tag_table",
				Fields: map[string]*orm.Field{
					"FirstName": {
						Column: "first_name_t",
					},
				},
			},
		},
		{
			name: "empty column",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column="`
				}
				return &TagTable{}
			}(),
			wantModel: &orm.Model{
				TableName: "tag_table",
				Fields: map[string]*orm.Field{
					"FirstName": {
						Column: "first_name",
					},
				},
			},
		},
		{
			name: "column only",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column"`
				}
				return &TagTable{}
			}(),
			wantErr: errs.NewErrInvalidTagContent("column"),
		},
		{
			name:   "table name",
			entity: &CustomTableName{},
			wantModel: &orm.Model{
				TableName: "custom_table_name_t",
				Fields: map[string]*orm.Field{
					"FirstName": {
						Column: "first_name",
					},
				},
			},
		},
		{
			name:   "table name ptr",
			entity: &CustomTableNamePtr{},
			wantModel: &orm.Model{
				TableName: "custom_table_name_ptr_t",
				Fields: map[string]*orm.Field{
					"FirstName": {
						Column: "first_name",
					},
				},
			},
		},
	}
	register := orm.NewRegistry()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := register.Get(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantModel, m)

			typ := reflect.TypeOf(tc.entity)
			cacheModel, ok := register.Models.Load(typ)
			assert.True(t, ok)
			assert.Equal(t, tc.wantModel, cacheModel)
		})
	}
}

type CustomTableName struct {
	FirstName string
}

func (c CustomTableName) TableName() string {
	return "custom_table_name_t"
}

type CustomTableNamePtr struct {
	FirstName string
}

func (c *CustomTableNamePtr) TableName() string {
	return "custom_table_name_ptr_t"
}

func TestModelWithTableName(t *testing.T) {
	registry := orm.NewRegistry()
	m, err := registry.Register(&TestModel{}, orm.ModelWithTableName("test_model_ttt"))
	require.NoError(t, err)
	assert.Equal(t, "test_model_ttt", m.TableName)
}

func TestModelWithColumnName(t *testing.T) {
	testCases := []struct {
		name        string
		field       string
		colName     string
		wantColName string
		wantErr     error
	}{
		{
			name:        "column name",
			field:       "FirstName",
			colName:     "first_name_ccc",
			wantColName: "first_name_ccc",
		},
		{
			name:    "invalid column name",
			field:   "xxx",
			colName: "first_name_ccc",
			wantErr: errs.NewErrUnknownField("xxx"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			registry := orm.NewRegistry()
			model, err := registry.Register(&TestModel{}, orm.ModelWithColumnName(tc.field, tc.colName))
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			fd, ok := model.Fields[tc.field]
			require.True(t, ok)
			assert.Equal(t, tc.wantColName, fd.Column)
		})
	}

}
