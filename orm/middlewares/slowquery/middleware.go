package querylog

import (
	"context"
	"github.com/dongma/imola/orm"
	"log"
	"time"
)

type MiddlewareBuilder struct {
	// 慢查询阀值
	threshold time.Duration
	logFunc   func(query string, args []any)
}

// NewMiddlewareBuilder 默认情况下threshold的值是100ms
func NewMiddlewareBuilder(threshold time.Duration) *MiddlewareBuilder {
	return &MiddlewareBuilder{
		logFunc: func(query string, args []any) {
			log.Printf("sql: %s, args: %v", query, args)
		},
		threshold: threshold,
	}
}

func (m *MiddlewareBuilder) LogFunc(fn func(query string, args []any)) *MiddlewareBuilder {
	m.logFunc = fn
	return m
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
