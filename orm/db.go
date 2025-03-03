package orm

import (
	"context"
	"database/sql"
	"imola/orm/internal/errs"
	"imola/orm/internal/valuer"
	"imola/orm/model"
)

type DBOption func(db *DB)

// DB DB是一个sql.DB的装饰器
type DB struct {
	core
	db *sql.DB
}

func Open(driver string, datasourceName string, opts ...DBOption) (*DB, error) {
	db, err := sql.Open(driver, datasourceName)
	if err != nil {
		return nil, err
	}
	return OpenDB(db, opts...)
}

func OpenDB(db *sql.DB, opts ...DBOption) (*DB, error) {
	res := &DB{
		core: core{
			r:       model.NewRegistry(),
			creator: valuer.NewUnsafeValue,
			dialect: DialectMySQL,
		},
		db: db,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func DBWithMiddlewares(mdls ...Middleware) DBOption {
	return func(db *DB) {
		db.mdls = mdls
	}
}

func DBWithDialect(dialect Dialect) DBOption {
	return func(db *DB) {
		db.dialect = dialect
	}
}

func DBUseReflect() DBOption {
	return func(db *DB) {
		db.creator = valuer.NewReflectValue
	}
}

func MustOpen(driver string, datasourceName string, opts ...DBOption) *DB {
	res, err := Open(driver, datasourceName, opts...)
	if err != nil {
		panic(err)
	}
	return res
}

type txKey struct{}

func (db *DB) BeginTxV2(ctx context.Context, opts *sql.TxOptions) (context.Context, *Tx, error) {
	val := ctx.Value(txKey{})
	tx, ok := val.(*Tx)
	// 存在一个事物，并且这个事务没有被提交或者回滚
	if ok && !tx.done {
		return ctx, tx, nil
	}
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	ctx = context.WithValue(ctx, txKey{}, tx)
	return ctx, tx, nil
}

// BeginTx 开启事务
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx}, nil
}

func (db *DB) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.db.QueryContext(ctx, query, args...)
}

func (db *DB) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.db.ExecContext(ctx, query, args...)
}

func (db *DB) getCore() core {
	return db.core
}

func (db *DB) DoTx(ctx context.Context,
	fn func(ctx context.Context, tx *Tx) error, opts *sql.TxOptions) (err error) {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	panicked := true
	defer func() {
		if panicked || err != nil {
			e := tx.Rollback()
			err = errs.NewErrFailedToRollbackTx(err, e, panicked)
		} else {
			err = tx.Commit()
		}
	}()
	err = fn(ctx, tx)
	panicked = false
	return err
}
