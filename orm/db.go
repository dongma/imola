package orm

import "database/sql"

type DBOption func(db *DB)

// DB DB是一个sql.DB的装饰器
type DB struct {
	r  *Registry
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
		r:  NewRegistry(),
		db: db,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func MustOpen(driver string, datasourceName string, opts ...DBOption) *DB {
	res, err := Open(driver, datasourceName, opts...)
	if err != nil {
		panic(err)
	}
	return res
}
