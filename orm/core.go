package orm

import (
	"context"
	"github.com/dongma/imola/orm/internal/valuer"
	"github.com/dongma/imola/orm/model"
	"github.com/dongma/imola/orm/sql"
)

type core struct {
	model   *model.Model
	dialect Dialect
	creator valuer.Creator
	r       *model.Registry
	mdls    []Middleware
}

func get[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	var root Handler = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return getHandler[T](ctx, sess, c, qc)
	}
	// 将ORM的中间件串起来, root指向的Handler最后执行
	for i := len(c.mdls) - 1; i >= 0; i-- {
		root = c.mdls[i](root)
	}
	return root(ctx, qc)
}

func getHandler[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	query, err := qc.Builder.Build()
	// 这个是构造SQL的报错
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}
	// 在这里发起查询，处理结果集合
	rows, err := sess.queryContext(ctx, query.SQL, query.Args)
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}
	// 未查询到结果数据时, 抛出异常 ErrorNoRows
	if !rows.Next() {
		return &QueryResult{
			Err: sql.ErrorNoRows,
		}
	}
	tp := new(T)
	val := c.creator(c.model, tp)
	err = val.SetColumn(rows)
	return &QueryResult{
		Err:    err,
		Result: tp,
	}
}

func exec(ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	query, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{
			Err: err,
			Result: sql.Result{
				Err: err,
			},
		}
	}
	res, err := sess.execContext(ctx, query.SQL, query.Args...)
	return &QueryResult{
		Err: err,
		Result: sql.Result{
			Err: err,
			Res: res,
		},
	}
}
