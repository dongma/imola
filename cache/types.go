package cache

import (
	"context"
	"time"
)

// Cache 屏蔽不同缓存中间件的差异
type Cache interface {
	// Get 按key从缓存中取出对应的值
	Get(ctx context.Context, key string) (any, error)

	// Set 指定key，在一定expiration内缓存value值
	Set(ctx context.Context, key string, value any, expiration time.Duration) error

	// Delete 缓存中删除指定的key
	Delete(ctx context.Context, key string) error

	// LoadAndDelete 缓存中加载key并返回，并删除key
	LoadAndDelete(ctx context.Context, key string) (any, error)
}
