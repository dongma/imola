package cache

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/singleflight"
	"log"
	"time"
)

var (
	ErrFailedToRefreshCache = errors.New("刷新缓存失败")
)

// ReadThroughCache 一定要赋值LoadFunc和Expiration，Expiration是过期时间
type ReadThroughCache struct {
	Cache
	Expiration time.Duration
	LoadFunc   func(ctx context.Context, key string) (any, error)
	g          singleflight.Group
}

// Get 普通操作，当val不存在时，调用loadFunc向cache加载数据
func (r *ReadThroughCache) Get(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if errors.Is(err, ErrKeyNotFound) {
		val, err = r.LoadFunc(ctx, key)
		if err == nil {
			er := r.Cache.Set(ctx, key, val, r.Expiration)
			if er != nil {
				return val, fmt.Errorf("%w, 原因: %s", ErrFailedToRefreshCache, er.Error())
			}
		}
	}
	return val, err
}

// GetAsync 当val不存在时，全异步向cache加载数据
func (r *ReadThroughCache) GetAsync(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if errors.Is(err, ErrKeyNotFound) {
		go func() {
			val, err = r.LoadFunc(ctx, key)
			if err == nil {
				er := r.Cache.Set(ctx, key, val, r.Expiration)
				if er != nil {
					log.Fatal(er)
				}
			}
		}()
	}
	return val, err
}

// SFGet SingleFlightGet
func (r *ReadThroughCache) SFGet(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if errors.Is(err, ErrKeyNotFound) {
		val, err, _ = r.g.Do(key, func() (interface{}, error) {
			cval, cerr := r.LoadFunc(ctx, key)
			if cerr == nil {
				er := r.Cache.Set(ctx, key, cval, r.Expiration)
				if er != nil {
					return cval, fmt.Errorf("%w, 原因: %s", ErrFailedToRefreshCache, er.Error())
				}
			}
			return cval, cerr
		})
	}
	return val, err
}
