package test

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"imola/orm/internal/errs"
	"imola/orm/model"
	"reflect"
	"testing"
)

func TestParseModel(t *testing.T) {
	testCases := []struct {
		name      string
		entity    any
		wantModel *model.Model
		wantErr   error
		fields    []*model.Field
		opts      []model.ModelOpt
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
			wantModel: &model.Model{
				TableName: "test_model",
			},
			fields: []*model.Field{
				{
					Column: "id",
					GoName: "Id",
					Typ:    reflect.TypeOf(int64(0)),
				},
				{
					Column: "first_name",
					GoName: "FirstName",
					Typ:    reflect.TypeOf(&sql.NullString{}),
					Offset: 8,
				},
				{
					Column: "last_name",
					GoName: "LastName",
					Typ:    reflect.TypeOf(&sql.NullString{}),
					Offset: 24,
				},
				{
					Column: "age",
					GoName: "Age",
					Typ:    reflect.TypeOf(int8(0)),
					Offset: 32,
				},
			},
		},
	}

	register := model.Registry{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := register.Register(tc.entity, tc.opts...)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			fieldMap := make(map[string]*model.Field)
			columnMap := make(map[string]*model.Field)
			for _, field := range tc.fields {
				fieldMap[field.GoName] = field
				columnMap[field.Column] = field
			}
			tc.wantModel.FieldMap = fieldMap
			tc.wantModel.ColumnMap = columnMap
			assert.Equal(t, tc.wantModel, m)
		})
	}
}

func TestRegistry_get(t *testing.T) {
	testCases := []struct {
		name      string
		entity    any
		wantModel *model.Model
		wantErr   error
		fields    []*model.Field
	}{
		{
			name:   "pointer",
			entity: &TestModel{},
			wantModel: &model.Model{
				TableName: "test_model",
			},
			fields: []*model.Field{
				{
					Column: "id",
					GoName: "Id",
					Typ:    reflect.TypeOf(int64(0)),
				},
				{
					Column: "first_name",
					GoName: "FirstName",
					Typ:    reflect.TypeOf(&sql.NullString{}),
					Offset: 8,
				},
				{
					Column: "last_name",
					GoName: "LastName",
					Typ:    reflect.TypeOf(&sql.NullString{}),
					Offset: 24,
				},
				{
					Column: "age",
					GoName: "Age",
					Typ:    reflect.TypeOf(int8(0)),
					Offset: 32,
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
			wantModel: &model.Model{
				TableName: "tag_table",
			},
			fields: []*model.Field{
				{
					Column: "first_name_t",
					GoName: "FirstName",
					Typ:    reflect.TypeOf(""),
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
			wantModel: &model.Model{
				TableName: "tag_table",
			},
			fields: []*model.Field{
				{
					Column: "first_name",
					GoName: "FirstName",
					Typ:    reflect.TypeOf(""),
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
			wantModel: &model.Model{
				TableName: "custom_table_name_t",
			},
			fields: []*model.Field{
				{
					Column: "first_name",
					GoName: "FirstName",
					Typ:    reflect.TypeOf(""),
				},
			},
		},
		{
			name:   "table name ptr",
			entity: &CustomTableNamePtr{},
			wantModel: &model.Model{
				TableName: "custom_table_name_ptr_t",
			},
			fields: []*model.Field{
				{
					Column: "first_name",
					GoName: "FirstName",
					Typ:    reflect.TypeOf(""),
				},
			},
		},
	}
	register := model.NewRegistry()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := register.Get(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}

			fieldMap := make(map[string]*model.Field)
			columnMap := make(map[string]*model.Field)
			for _, field := range tc.fields {
				fieldMap[field.GoName] = field
				columnMap[field.Column] = field
			}
			tc.wantModel.FieldMap = fieldMap
			tc.wantModel.ColumnMap = columnMap
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
	registry := model.NewRegistry()
	m, err := registry.Register(&TestModel{}, model.ModelWithTableName("test_model_ttt"))
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
			registry := model.NewRegistry()
			model, err := registry.Register(&TestModel{}, model.ModelWithColumnName(tc.field, tc.colName))
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			fd, ok := model.FieldMap[tc.field]
			require.True(t, ok)
			assert.Equal(t, tc.wantColName, fd.Column)
		})
	}

}
