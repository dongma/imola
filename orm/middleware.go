package orm

import (
	"context"
	"imola/orm/model"
)

type QueryContext struct {
	// 查询类型，记录增删改查
	Type string
	// 代表的是查询本身
	Builder QueryBuilder
	Model   *model.Model
}

type QueryResult struct {
	// Result 在不同查询下类型是不同的，SELECT可以是*T，也可以是[]*T
	Result any
	// 查询本身出的问题
	Err error
}

type Handler func(ctx context.Context, qc *QueryContext) *QueryResult

// Middleware 实现AOP，将中间件一个一个串起来
type Middleware func(next Handler) Handler
