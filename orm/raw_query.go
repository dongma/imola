package orm

import (
	"context"
	"database/sql"
	sql2 "github.com/dongma/imola/orm/sql"
)

type RawQuerier[T any] struct {
	core
	sess Session
	sql  string
	args []any
}

func (r *RawQuerier[T]) Build() (*Query, error) {
	return &Query{
		SQL:  r.sql,
		Args: r.args,
	}, nil
}

func RawQuery[T any](sess Session, sql string, args ...interface{}) *RawQuerier[T] {
	c := sess.getCore()
	return &RawQuerier[T]{
		sess: sess,
		sql:  sql,
		args: args,
		core: c,
	}
}

func (r *RawQuerier[T]) Exec(ctx context.Context) sql2.Result {
	var err error
	r.model, err = r.r.Get(new(T))
	if err != nil {
		return sql2.Result{
			Err: err,
		}
	}
	res := exec(ctx, r.sess, r.core, &QueryContext{
		Type:    "RAW",
		Builder: r,
		Model:   r.model,
	})

	var sqlRes sql.Result
	return sql2.Result{
		Err: res.Err,
		Res: sqlRes,
	}
}

func (r *RawQuerier[T]) Get(ctx context.Context) (*T, error) {
	var err error
	r.model, err = r.r.Get(new(T))
	if err != nil {
		return nil, err
	}
	res := get[T](ctx, r.sess, r.core, &QueryContext{
		Type:    "RAW",
		Builder: r,
		Model:   r.model,
	})
	if res.Result != nil {
		return res.Result.(*T), res.Err
	}
	return nil, res.Err
}

func (r *RawQuerier[T]) GetMulti(ctx context.Context) ([]*T, error) {
	panic("not implemented")
}
