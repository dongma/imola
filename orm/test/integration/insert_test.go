//go:build e2e

package integration

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"imola/orm"
	"testing"
	"time"
)

type InsertSuite struct {
	Suite
}

func TestMysqlInsert(t *testing.T) {
	suite.Run(t, &InsertSuite{
		Suite{
			driver: "mysql",
			dsn:    "root:root@tcp(localhost:13306)/integration_test",
		},
	})
}

func (i *InsertSuite) TestInsert() {
	db := i.db
	t := i.T()
	testCases := []struct {
		name         string
		i            *orm.Inserter[SimpleStruct]
		wantAffected int64 // 插入行数
	}{
		{
			name:         "insert one",
			i:            orm.NewInserter[SimpleStruct](db).Values(NewSimpleStruct(10)),
			wantAffected: 1,
		},
		{
			name: "insert multiple",
			i: orm.NewInserter[SimpleStruct](db).Values(
				NewSimpleStruct(12),
				NewSimpleStruct(13)),
			wantAffected: 2,
		},
		{
			name:         "insert id",
			i:            orm.NewInserter[SimpleStruct](db).Values(&SimpleStruct{Id: 15}),
			wantAffected: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			res := tc.i.Exec(ctx)
			affected, err := res.RowsAffected()
			assert.NoError(t, err)
			assert.Equal(t, tc.wantAffected, affected)
		})
	}
}
