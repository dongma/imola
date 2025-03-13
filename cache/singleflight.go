package cache

import (
	"context"
	"golang.org/x/sync/singleflight"
	"time"
)

// SingleflightCache 使用singleflight减轻数据库压力，多个goroutine试图到数据库中加载
// 同一个key时，只允许一个gouroutine去查询，其它都在原地等待结果。
type SingleflightCache struct {
	ReadThroughCache
}

func NewSingleflightCache(cache Cache,
	loadFunc func(ctx context.Context, key string) (any, error),
	expiration time.Duration) *SingleflightCache {
	g := &singleflight.Group{}
	return &SingleflightCache{
		ReadThroughCache: ReadThroughCache{
			Cache: cache,
			LoadFunc: func(ctx context.Context, key string) (any, error) {
				val, err, _ := g.Do(key, func() (interface{}, error) {
					return loadFunc(ctx, key)
				})
				return val, err
			},
			Expiration: expiration,
		},
	}
}
