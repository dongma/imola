package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	errKeyNotFound = errors.New("cache: 键不存在")
	errKeyExpired  = errors.New("cache: 键过期")
)

type MemoryCache struct {
	data  map[string]*item
	mutex sync.RWMutex
	close chan struct{}
}

func NewMemoryCache(interval time.Duration) *MemoryCache {
	res := &MemoryCache{
		data: make(map[string]*item),
	}
	go func() {
		ticker := time.NewTicker(interval)
		counter := 0
		for {
			select {
			case t := <-ticker.C:
				res.mutex.Lock()
				for key, val := range res.data {
					if counter > 1000 {
						break
					}
					if !val.deadline.IsZero() && val.deadline.Before(t) {
						delete(res.data, key)
					}
					counter++
				}
				res.mutex.Unlock()
			case <-res.close:
				return
			}
		}
	}()
	return res
}

func (m *MemoryCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	var dl time.Time
	if expiration > 0 {
		dl = time.Now().Add(expiration)
	}
	m.data[key] = &item{
		val:      value,
		deadline: dl,
	}

	// case1: 每个key一个goroutine, 当key过期的时，执行删除操作
	/*if expiration > 0 {
		time.AfterFunc(expiration, func() {
			m.mutex.Lock()
			defer m.mutex.Unlock()
			val, ok := m.data[key]
			// 判断条件，key存在 && key设置了超时时间 && key已过期
			if ok && !val.deadline.IsZero() && val.deadline.Before(time.Now()) {
				delete(m.data, key)
			}
		})
	}*/
	return nil
}

func (m *MemoryCache) Get(ctx context.Context, key string) (any, error) {
	m.mutex.RLock()
	res, ok := m.data[key]
	m.mutex.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: key: %s", errKeyNotFound, key)
	}
	now := time.Now()
	if res.deadlineBefore(now) {
		m.mutex.Lock()
		defer m.mutex.Unlock()
		// 进行了 double check
		res, ok = m.data[key]
		if !ok {
			return nil, fmt.Errorf("%w: key: %s", errKeyNotFound, key)
		}
		if res.deadlineBefore(now) {
			delete(m.data, key)
			return nil, fmt.Errorf("%w: key: %s", errKeyExpired, key)
		}
	}
	return res.val, nil
}

func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.data, key)
	return nil
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
