package test

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"imola/orm"
	"imola/orm/internal/errs"
	"testing"
)

func TestSelector_Build(t *testing.T) {
	db, err := orm.NewDB()
	require.NoError(t, err)
	testCases := []struct {
		name      string
		builder   orm.QueryBuilder
		wantQuery *orm.Query
		wantErr   error
	}{
		// where部分的测试，支持AND、NOT、OR类型
		{
			name:    "empty where",
			builder: orm.NewSelector[TestModel](db).Where(),
			wantQuery: &orm.Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},
		{
			name:    "where",
			builder: orm.NewSelector[TestModel](db).Where(orm.C("Age").Eq(18)),
			wantQuery: &orm.Query{
				SQL:  "SELECT * FROM `test_model` WHERE `age` = ?;",
				Args: []any{18},
			},
		},
		{
			name:    "not",
			builder: orm.NewSelector[TestModel](db).Where(orm.Not(orm.C("Age").Eq(18))),
			wantQuery: &orm.Query{
				SQL:  "SELECT * FROM `test_model` WHERE  NOT (`age` = ?);",
				Args: []any{18},
			},
		},
		{
			name:    "and",
			builder: orm.NewSelector[TestModel](db).Where(orm.C("Age").Eq(18).And(orm.C("FirstName").Eq("Tom"))),
			wantQuery: &orm.Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` = ?) AND (`first_name` = ?);",
				Args: []any{18, "Tom"},
			},
		},
		{
			name:    "or",
			builder: orm.NewSelector[TestModel](db).Where(orm.C("Age").Eq(18).Or(orm.C("FirstName").Eq("Tom"))),
			wantQuery: &orm.Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` = ?) OR (`first_name` = ?);",
				Args: []any{18, "Tom"},
			},
		},
		{
			name:    "invalid column",
			builder: orm.NewSelector[TestModel](db).Where(orm.C("Age").Eq(18).Or(orm.C("xxxx").Eq("Tom"))),
			wantErr: errs.NewErrUnknownField("xxxx"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.builder.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
