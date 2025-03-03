package querylog

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"imola/orm"
	"testing"
)

func TestMiddlewareBuilder_Build(t *testing.T) {
	var query string
	var args []any
	m := (&MiddlewareBuilder{}).LogFunc(func(q string, ags []any) {
		query = q
		args = ags
	})

	db, err := orm.Open("sqlite3", "file:test.db?cache=shared&mode=memory",
		orm.DBWithMiddlewares(m.Build()))
	require.NoError(t, err)
	orm.NewSelector[TestModel](db).Where(orm.C("Id").Eq(10)).Get(context.Background())
	assert.Equal(t, "SELECT * FROM `test_model` WHERE `id` = ?;", query)
	assert.Equal(t, []any{10}, args)

	orm.NewInserter[TestModel](db).Values(&TestModel{Id: 18}).Exec(context.Background())
	assert.Equal(t, "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?);", query)
	assert.Equal(t, []any{int64(18), "", int8(0), (*sql.NullString)(nil)}, args)
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
