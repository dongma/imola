package orm

import (
	"context"
	"database/sql"
)

var (
	_ Session = &Tx{}
	_ Session = &DB{}
)

// Session 代表一个DB和Tx的公共抽象，在ORM语境下，一般代表一个上下文，也可以理解为一种分组机制（在这个组内共享一些配置）
type Session interface {
	getCore() core
	queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	execContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type Tx struct {
	tx *sql.Tx
	db *DB
	// 给事务扩展方案
	done bool
}

func (t *Tx) getCore() core {
	return t.db.core
}

func (t *Tx) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (t *Tx) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *Tx) Commit() error {
	t.done = true
	return t.tx.Commit()
}

func (t *Tx) Rollback() error {
	t.done = true
	return t.tx.Rollback()
}

func (t *Tx) RollbackIfNotCommit() error {
	t.done = true
	err := t.tx.Rollback()
	if err == sql.ErrTxDone {
		return nil
	}
	return err
}
