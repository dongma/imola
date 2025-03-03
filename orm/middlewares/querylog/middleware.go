package querylog

import (
	"context"
	"imola/orm"
	"log"
)

type MiddlewareBuilder struct {
	logFunc func(query string, args []any)
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{
		logFunc: func(query string, args []any) {
			log.Printf("sql: %s, args: %v", query, args)
		},
	}
}

func (m *MiddlewareBuilder) LogFunc(fn func(query string, args []any)) *MiddlewareBuilder {
	m.logFunc = fn
	return m
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			query, err := qc.Builder.Build()
			if err != nil {
				return &orm.QueryResult{
					Err: err,
				}
			}
			m.logFunc(query.SQL, query.Args)
			res := next(ctx, qc)
			return res
		}
	}
}
