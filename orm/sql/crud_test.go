package sql

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
	"time"
)

func TestDB(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	// 除了sql语句， 都是使用ExecContext来执行
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS test_model(
		    id INTEGER PRIMARY KEY,
		    first_name text NOT NULL,
		    age INTEGER,
		    last_name text NOT NULL
		)
	`)
	// 完成了数据表的创建
	require.NoError(t, err)
	// 使用?作为查询参数的占位符
	res, err := db.ExecContext(ctx, "INSERT INTO `test_model` VALUES(?, ?, ?, ?)",
		1, "Tom", 18, "Jerry")
	require.NoError(t, err)
	affected, err := res.RowsAffected()
	require.NoError(t, err)
	log.Println("受影响的行数:", affected)
	lastId, err := res.LastInsertId()
	require.NoError(t, err)
	log.Println("最后插入的id:", lastId)

	// 执行查询语句
	row := db.QueryRowContext(ctx,
		"select id, first_name, age, last_name from `test_model` where id = ?", 1)
	require.NoError(t, row.Err())
	tm := TestModel{}
	err = row.Scan(&tm.Id, &tm.FirstName, &tm.Age, &tm.LastName)
	require.NoError(t, err)

	row = db.QueryRowContext(ctx,
		"select id, first_name, age, last_name from `test_model` where id = ?", 2)
	// 查询不到数据
	require.Error(t, sql.ErrNoRows, row.Err())

	cancel()
}

func TestPreparedStatement(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS test_model(
		    id INTEGER PRIMARY KEY,
		    first_name text NOT NULL,
		    age INTEGER,
		    last_name text NOT NULL
		)
	`)
	// 完成了数据表的创建
	require.NoError(t, err)
	defer db.Close()

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	stmt, err := db.PrepareContext(ctx, "select * from `test_model` where id = ?")
	require.NoError(t, err)
	// id=1
	rows, err := stmt.QueryContext(ctx, 1)
	for rows.Next() {
		tm := &TestModel{}
		err = rows.Scan(&tm.Id, &tm.FirstName, &tm.Age, &tm.LastName)
		require.NoError(t, err)
		log.Println(tm)
	}
	cancel()
	// 整个应用关闭时调用
	db.Close()
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
