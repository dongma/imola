package cache

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

var (
	errOverCapacity = errors.New("cache: 超过容量限制")
)

// MaxCntCache 控制住缓存住的键值对的数量
type MaxCntCache struct {
	*MemoryCache
	cnt    int32
	maxCnt int32
}

func NewMaxCntCache(mcache *MemoryCache, maxCnt int32) *MaxCntCache {
	res := &MaxCntCache{
		MemoryCache: mcache,
		maxCnt:      maxCnt,
	}
	origin := mcache.onEvicted

	res.onEvicted = func(key string, val any) {
		atomic.AddInt32(&res.cnt, -1)
		if origin != nil {
			origin(key, val)
		}
	}
	return res
}

func (c *MaxCntCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	_, ok := c.Data[key]
	if !ok {
		if c.cnt+1 > c.maxCnt {
			// 后面，你可以在这里设计复杂的淘汰策略
			return errOverCapacity
		}
		c.cnt++
	}
	return c.SetVal(key, val, expiration)
}
