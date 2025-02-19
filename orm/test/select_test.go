package test

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"imola/orm"
	"testing"
)

func TestSelector_Build(t *testing.T) {
	testCases := []struct {
		name      string
		builder   orm.QueryBuilder
		wantQuery *orm.Query
		wantErr   error
	}{
		// where部分的测试，支持AND、NOT、OR类型
		{
			name:    "empty where",
			builder: (&orm.Selector[TestModel]{}).Where(),
			wantQuery: &orm.Query{
				SQL: "SELECT * FROM `TestModel`;",
			},
		},
		{
			name:    "where",
			builder: (&orm.Selector[TestModel]{}).Where(orm.C("Age").Eq(18)),
			wantQuery: &orm.Query{
				SQL:  "SELECT * FROM `TestModel` WHERE `Age` = ?;",
				Args: []any{18},
			},
		},
		{
			name:    "not",
			builder: (&orm.Selector[TestModel]{}).Where(orm.Not(orm.C("Age").Eq(18))),
			wantQuery: &orm.Query{
				SQL:  "SELECT * FROM `TestModel` WHERE  NOT (`Age` = ?);",
				Args: []any{18},
			},
		},
		{
			name:    "and",
			builder: (&orm.Selector[TestModel]{}).Where(orm.C("Age").Eq(18).And(orm.C("FirstName").Eq("Tom"))),
			wantQuery: &orm.Query{
				SQL:  "SELECT * FROM `TestModel` WHERE (`Age` = ?) AND (`FirstName` = ?);",
				Args: []any{18, "Tom"},
			},
		},
		{
			name:    "or",
			builder: (&orm.Selector[TestModel]{}).Where(orm.C("Age").Eq(18).Or(orm.C("FirstName").Eq("Tom"))),
			wantQuery: &orm.Query{
				SQL:  "SELECT * FROM `TestModel` WHERE (`Age` = ?) OR (`FirstName` = ?);",
				Args: []any{18, "Tom"},
			},
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
