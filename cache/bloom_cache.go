package cache

import (
	"context"
)

// BloomFilterCache 为了解决缓存穿透，BloomFilter认为key1存在时，才会最终去数据库中查询
type BloomFilterCache struct {
	ReadThroughCache
}

func NewBloomFilterCache(cache Cache, bf BloomFilter,
	loadFunc func(ctx context.Context, key string) (any, error)) *BloomFilterCache {
	return &BloomFilterCache{
		ReadThroughCache: ReadThroughCache{
			Cache: cache,
			LoadFunc: func(ctx context.Context, key string) (any, error) {
				if !bf.HasKey(ctx, key) {
					return nil, ErrKeyNotFound
				}
				return loadFunc(ctx, key)
			},
		},
	}
}

// BloomFilter Bloom过滤器，用于快速检查key在缓存中是否存在
type BloomFilter interface {
	HasKey(ctx context.Context, key string) bool
}
