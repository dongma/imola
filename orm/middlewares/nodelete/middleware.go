package querylog

import (
	"context"
	"imola/orm"
	"time"
)

type MiddlewareBuilder struct {
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{}
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			startTime := time.Now()
			defer func() {
				duration := time.Since(startTime)
				if duration <= m.threshold {
					return
				}
				query, err := qc.Builder.Build()
				if err == nil {
					m.logFunc(query.SQL, query.Args)
				}
			}()

			return next(ctx, qc)
		}
	}
}
