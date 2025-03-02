package test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"imola/orm"
	"imola/orm/internal/errs"
	"testing"
)

func TestSelector_Build(t *testing.T) {
	db := memoryDB(t, orm.DBWithDialect(orm.DialectMySQL))
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
		{
			name:    "avg",
			builder: orm.NewSelector[TestModel](db).Select(orm.Avg("Age")),
			wantQuery: &orm.Query{
				SQL: "SELECT AVG(`age`) FROM `test_model`;",
			},
		},
		{
			name:    "avg alias",
			builder: orm.NewSelector[TestModel](db).Select(orm.Avg("Age").As("avg_age")),
			wantQuery: &orm.Query{
				SQL: "SELECT AVG(`age`) AS `avg_age` FROM `test_model`;",
			},
		},
		{
			name:    "sum",
			builder: orm.NewSelector[TestModel](db).Select(orm.Sum("Age")),
			wantQuery: &orm.Query{
				SQL: "SELECT SUM(`age`) FROM `test_model`;",
			},
		},
		{
			name:    "raw expr",
			builder: orm.NewSelector[TestModel](db).Select(orm.Raw("COUNT(DISTINCT `first_name`)")),
			wantQuery: &orm.Query{
				SQL: "SELECT COUNT(DISTINCT `first_name`) FROM `test_model`;",
			},
		},
		{
			name:    "raw expr as predicate",
			builder: orm.NewSelector[TestModel](db).Where(orm.Raw("`id` < ?", 18).AsPredicate()),
			wantQuery: &orm.Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`id` < ?);",
				Args: []any{18},
			},
		},
		{
			name:    "raw expr used in predicate",
			builder: orm.NewSelector[TestModel](db).Where(orm.C("Id").Eq(orm.Raw("`age` + ?", 1))),
			wantQuery: &orm.Query{
				SQL:  "SELECT * FROM `test_model` WHERE `id` = (`age` + ?);",
				Args: []any{1},
			},
		},
		{
			name:    "column alias",
			builder: orm.NewSelector[TestModel](db).Select(orm.C("FirstName").As("my_name"), orm.C("LastName").As("")),
			wantQuery: &orm.Query{
				SQL: "SELECT `first_name` AS `my_name`,`last_name` FROM `test_model`;",
			},
		},
		{
			name:    "column alias in where",
			builder: orm.NewSelector[TestModel](db).Where(orm.C("Id").As("my_id").Eq(18)),
			wantQuery: &orm.Query{
				SQL:  "SELECT * FROM `test_model` WHERE `id` = ?;",
				Args: []any{18},
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

func TestSelector_Get(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db, err := orm.OpenDB(mockDB)
	require.NoError(t, err)
	// 对应于query error
	mock.ExpectQuery("SELECT .*").WillReturnError(errors.New("query error"))
	// 对应于no rows
	rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	mock.ExpectQuery("SELECT .* WHERE `id` <.").WillReturnRows(rows)
	// 对应于data
	rows = sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	rows.AddRow("1", "Tom", "18", "Jerry")
	mock.ExpectQuery("SELECT .* WHERE `id` =.").WillReturnRows(rows)

	testCases := []struct {
		name    string
		s       *orm.Selector[TestModel]
		wantErr error
		wantRes *TestModel
	}{
		{
			name:    "invalid query",
			s:       orm.NewSelector[TestModel](db).Where(orm.C("xxx").Eq(18)),
			wantErr: errs.NewErrUnknownField("xxx"),
		},
		{
			name:    "query error",
			s:       orm.NewSelector[TestModel](db).Where(orm.C("Id").Eq(1)),
			wantErr: errors.New("query error"),
		},
		{
			name:    "no rows",
			s:       orm.NewSelector[TestModel](db).Where(orm.C("Id").Lt(1)),
			wantErr: orm.ErrorNoRows,
		},
		{
			name: "data",
			s:    orm.NewSelector[TestModel](db).Where(orm.C("Id").Eq(1)),
			wantRes: &TestModel{
				Id:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "Jerry"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.s.Get(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func memoryDB(t *testing.T, opts ...orm.DBOption) *orm.DB {
	db, err := orm.Open("sqlite3", "file:test.db?cache=shared&mode=memory",
		opts...)
	require.NoError(t, err)
	return db
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
