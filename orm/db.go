package orm

import (
	"database/sql"
	"imola/orm/internal/valuer"
	"imola/orm/model"
)

type DBOption func(db *DB)

// DB DB是一个sql.DB的装饰器
type DB struct {
	r       *model.Registry
	db      *sql.DB
	creator valuer.Creator
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
		r:       model.NewRegistry(),
		db:      db,
		creator: valuer.NewUnsafeValue,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
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
