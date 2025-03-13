package cache

import (
	"context"
	"math/rand"
	"time"
)

// RandomExpirationCache 解决缓存雪崩，在设置key过期时，加上一个随机的偏移量
type RandomExpirationCache struct {
	Cache
}

func (r *RandomExpirationCache) Set(ctx context.Context, key string, val string,
	expiration time.Duration) error {
	if expiration > 0 {
		// 加上一个 [0,300)s的偏移量
		offset := time.Duration(rand.Intn(300)) * time.Second
		expiration = expiration + offset
		return r.Cache.Set(ctx, key, val, expiration)
	}
	return r.Cache.Set(ctx, key, val, expiration)
}
