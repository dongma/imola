package sql

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestSqlMock(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockRows := sqlmock.NewRows([]string{"id", "first_name"})
	mockRows.AddRow(1, "Tom")
	// .*实际上是正则表达式
	mock.ExpectQuery("select id, first_name from `user`.*").WillReturnRows(mockRows)
	mock.ExpectQuery("select id from `user`.*").WillReturnError(errors.New("mock error"))

	rows, err := db.QueryContext(context.Background(), "select id, first_name from `user` where id = 1")
	require.NoError(t, err)
	for rows.Next() {
		tm := &TestModel{}
		err = rows.Scan(&tm.Id, &tm.FirstName)
		require.NoError(t, err)
		log.Println(tm)
	}

	_, err = db.QueryContext(context.Background(), "select id from `user` where id = 1")
	require.Error(t, err)
}
