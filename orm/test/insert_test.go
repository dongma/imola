package test

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"imola/orm"
	"imola/orm/internal/errs"
	"testing"
)

func TestInserter_Build(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct {
		name      string
		i         orm.QueryBuilder
		wantErr   error
		wantQuery *orm.Query
	}{
		{
			// 只插入一行
			name:    "no row",
			i:       orm.NewInserter[TestModel](db).Values(),
			wantErr: errs.ErrInsertZeroRow,
		},
		{
			// 只插入一行
			name: "single row",
			i: orm.NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{String: "Jerry", Valid: true},
			}),
			wantQuery: &orm.Query{
				SQL:  "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?);",
				Args: []any{int64(12), "Tom", int8(18), &sql.NullString{String: "Jerry", Valid: true}},
			},
		},
		{
			// 插入多行
			name: "multi rows",
			i: orm.NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{String: "Jerry", Valid: true},
			}, &TestModel{
				Id:        13,
				FirstName: "DaMing",
				Age:       19,
				LastName:  &sql.NullString{String: "Deng", Valid: true},
			}),
			wantQuery: &orm.Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?),(?,?,?,?);",
				Args: []any{int64(12), "Tom", int8(18), &sql.NullString{String: "Jerry", Valid: true},
					int64(13), "DaMing", int8(19), &sql.NullString{String: "Deng", Valid: true}},
			},
		},
		{
			// 插入多行
			name: "partial rows",
			i: orm.NewInserter[TestModel](db).Columns("Id", "FirstName").Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{String: "Jerry", Valid: true},
			}, &TestModel{
				Id:        13,
				FirstName: "DaMing",
				Age:       19,
				LastName:  &sql.NullString{String: "Deng", Valid: true},
			}),
			wantQuery: &orm.Query{
				SQL:  "INSERT INTO `test_model`(`id`,`first_name`) VALUES (?,?),(?,?);",
				Args: []any{int64(12), "Tom", int64(13), "DaMing"},
			},
		},
		{
			// 只插入一行
			name: "upsert-update value",
			i: orm.NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{String: "Jerry", Valid: true},
			}).OnDuplicateKey().Update(orm.Assign("FirstName", "Deng"),
				orm.Assign("Age", 19)),
			wantQuery: &orm.Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?) " +
					"ON DUPLICATE KEY UPDATE `first_name`=?,`age`=?;",
				Args: []any{int64(12), "Tom", int8(18), &sql.NullString{String: "Jerry", Valid: true}, "Deng", 19},
			},
		},
		{
			// 插入多行
			name: "upsert-update column",
			i: orm.NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{String: "Jerry", Valid: true},
			}, &TestModel{
				Id:        13,
				FirstName: "DaMing",
				Age:       19,
				LastName:  &sql.NullString{String: "Deng", Valid: true},
			}).OnDuplicateKey().Update(orm.C("FirstName"), orm.C("Age")),
			wantQuery: &orm.Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?),(?,?,?,?)" +
					" ON DUPLICATE KEY UPDATE `first_name`=VALUES(`first_name`),`age`=VALUES(`age`);",
				Args: []any{int64(12), "Tom", int8(18), &sql.NullString{String: "Jerry", Valid: true},
					int64(13), "DaMing", int8(19), &sql.NullString{String: "Deng", Valid: true}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.i.Build()
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}

}
