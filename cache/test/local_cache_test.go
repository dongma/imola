package test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"imola/cache"
	"testing"
	"time"
)

func TestBuildInMapCache_Get(t *testing.T) {
	testCases := []struct {
		name    string
		key     string
		cache   func() *cache.MemoryCache
		wantVal any
		wantErr error
	}{
		{
			name: "key not found",
			key:  "not exist key",
			cache: func() *cache.MemoryCache {
				return cache.NewMemoryCache(10 * time.Second)
			},
			wantErr: fmt.Errorf("%w, key: %s", cache.ErrKeyNotFound, "not exist key"),
		},
		{
			name: "get value",
			key:  "key1",
			cache: func() *cache.MemoryCache {
				res := cache.NewMemoryCache(10 * time.Second)
				err := res.Set(context.Background(), "key1", 123, time.Minute)
				require.NoError(t, err)
				return res
			},
			wantVal: 123,
		},
		{
			name: "expired",
			key:  "expired key",
			cache: func() *cache.MemoryCache {
				res := cache.NewMemoryCache(10 * time.Second)
				err := res.Set(context.Background(), "expired key", 123, time.Second)
				require.NoError(t, err)
				time.Sleep(2 * time.Second)
				return res
			},
			wantErr: fmt.Errorf("%w, key: %s", cache.ErrKeyNotFound, "expired key"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := tc.cache().Get(context.Background(), tc.key)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantVal, val)
		})
	}
}

func TestBuildInMapCache_Loop(t *testing.T) {
	cnt := 0
	mcache := cache.NewMemoryCache(time.Second, cache.WithEvictedCallbackOption(func(key string, val any) {
		cnt++
	}))
	err := mcache.Set(context.Background(), "key1", 123, time.Second)
	require.NoError(t, err)
	time.Sleep(time.Second * 3)
	mcache.Mutex.RLock()
	defer mcache.Mutex.RUnlock()
	_, ok := mcache.Data["key1"]
	require.False(t, ok)
	require.Equal(t, 1, cnt)
}
