package querylog

import (
	"context"
	"errors"
	"github.com/dongma/imola/orm"
	"strings"
)

// MiddlewareBuilder 要强制查询语句
// 1、select、update、delete必须要带where
// 2、update和delete必须要带where
type MiddlewareBuilder struct {
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{}
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			if qc.Type == "select" || qc.Type == "insert" {
				return next(ctx, qc)
			}
			query, err := qc.Builder.Build()
			if err != nil {
				return &orm.QueryResult{
					Err: err,
				}
			}
			if !strings.Contains(query.SQL, "WHERE") {
				return &orm.QueryResult{
					Err: errors.New("不准执行没有WHERE的delete或update语句"),
				}
			}
			return next(ctx, qc)
		}
	}
}
