package orm

import (
	"context"
	"database/sql"
)

// Querier 用于Select语句
type Querier[T any] interface {
	Get(ctx context.Context) (*T, error)
	GetMulti(ctx context.Context) ([]*T, error)
}

// Executor 用于Insert、Update和Delete
type Executor interface {
	Exec(ctx context.Context) (sql.Result, error)
}

type QueryBuilder interface {
	Build() (*Query, error)
}

type Query struct {
	SQL  string
	Args []any
}

// TableName 用户实现这个接口来返回自定义的表名
type TableName interface {
	TableName() string
}
