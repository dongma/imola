package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrKeyNotFound = errors.New("cache: 键不存在")
)

type MemoryCacheOption func(cache *MemoryCache)

type MemoryCache struct {
	Data      map[string]*item
	Mutex     sync.RWMutex
	close     chan struct{}
	onEvicted func(key string, val any)
}

func NewMemoryCache(interval time.Duration, opts ...MemoryCacheOption) *MemoryCache {
	res := &MemoryCache{
		Data:  make(map[string]*item, 100),
		close: make(chan struct{}),
		onEvicted: func(key string, val any) {
		},
	}

	for _, opt := range opts {
		opt(res)
	}

	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case t := <-ticker.C:
				res.Mutex.Lock()
				i := 0
				for key, val := range res.Data {
					if i > 1000 {
						break
					}
					if val.deadlineBefore(t) {
						res.delete(key)
					}
					i++
				}
				res.Mutex.Unlock()
			case <-res.close:
				return
			}
		}
	}()
	return res
}

func WithEvictedCallbackOption(fn func(key string, val any)) MemoryCacheOption {
	return func(cache *MemoryCache) {
		cache.onEvicted = fn
	}
}

func (m *MemoryCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	return m.SetVal(key, value, expiration)
}

func (m *MemoryCache) SetVal(key string, value any, expiration time.Duration) error {
	var dl time.Time
	if expiration > 0 {
		dl = time.Now().Add(expiration)
	}
	m.Data[key] = &item{
		val:      value,
		deadline: dl,
	}
	return nil
}

func (m *MemoryCache) Get(ctx context.Context, key string) (any, error) {
	m.Mutex.RLock()
	res, ok := m.Data[key]
	m.Mutex.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w, key: %s", ErrKeyNotFound, key)
	}
	now := time.Now()
	if res.deadlineBefore(now) {
		m.Mutex.Lock()
		defer m.Mutex.Unlock()
		// 进行了 double check
		res, ok = m.Data[key]
		if !ok {
			return nil, fmt.Errorf("%w, key: %s", ErrKeyNotFound, key)
		}
		if res.deadlineBefore(now) {
			m.delete(key)
			return nil, fmt.Errorf("%w, key: %s", ErrKeyNotFound, key)
		}
	}
	return res.val, nil
}

func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.delete(key)
	return nil
}

func (m *MemoryCache) delete(key string) {
	itm, ok := m.Data[key]
	if !ok {
		return
	}
	delete(m.Data, key)
	m.onEvicted(key, itm.val)
}

func (m *MemoryCache) Close() error {
	select {
	case m.close <- struct{}{}:
	default:
		return errors.New("重复关闭")
	}
	return nil
}

type item struct {
	val      any
	deadline time.Time
}

func (i *item) deadlineBefore(t time.Time) bool {
	return !i.deadline.IsZero() && i.deadline.Before(t)
}
