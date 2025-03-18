package orm

import (
	"context"
	"imola/orm/sql"
)

// Querier 用于Select语句
type Querier[T any] interface {
	Get(ctx context.Context) (*T, error)
	GetMulti(ctx context.Context) ([]*T, error)
}

// Executor 用于Insert、Update和Delete
type Executor interface {
	Exec(ctx context.Context) sql.Result
}

type QueryBuilder interface {
	Build() (*Query, error)
}

type Query struct {
	SQL  string
	Args []any
}
