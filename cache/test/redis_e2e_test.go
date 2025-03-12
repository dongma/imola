//go:build e2e

package test

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"imola/cache"
	"testing"
	"time"
)

func TestRedisCache_e2e_Get(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	rcache := cache.NewRedisCache(rdb)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err := rcache.Set(ctx, "key1", "value1", time.Minute)
	require.NoError(t, err)
	val, err := rcache.Get(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", val)
}
