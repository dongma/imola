package test

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"imola/cache/distlock"
	"testing"
	"time"
)

func TestClient_e2e_TryLock(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	testCases := []struct {
		name       string
		before     func(t *testing.T)
		after      func(t *testing.T)
		key        string
		expiration time.Duration
		wantErr    error
		wantLock   *distlock.Lock
	}{
		{
			// 别人持有锁了
			name: "key exist",
			before: func(t *testing.T) {
				// 模拟别人持有锁
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "key1", "value1", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.GetDel(ctx, "key1").Result()
				require.NoError(t, err)
				assert.Equal(t, "value1", res)
			},
			key:     "key1",
			wantErr: distlock.ErrFailedToPreemptLock,
			wantLock: &distlock.Lock{
				Key: "key1",
			},
		},
		{
			// 你加锁成功
			name: "locked",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.GetDel(ctx, "key2").Result()
				require.NoError(t, err)
				// 加锁成功意味着你应该设置好了值
				assert.NotEmpty(t, res)
			},
			key: "key2",
			wantLock: &distlock.Lock{
				Key: "key2",
			},
		},
	}

	client := distlock.NewClient(rdb)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			lock, err := client.TryLock(ctx, tc.key, time.Minute)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantLock.Key, lock.Key)
			assert.NotEmpty(t, lock.Value)
			assert.NotNil(t, lock.Client)
			tc.after(t)
		})
	}
}

func TestClient_e2e_UnLock(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	testCases := []struct {
		name    string
		before  func(t *testing.T)
		after   func(t *testing.T)
		lock    *distlock.Lock
		wantErr error
	}{
		{
			name: "lock not hold",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {

			},
			lock: &distlock.Lock{
				Key:    "unlock_key1",
				Value:  "123",
				Client: rdb,
			},
			wantErr: distlock.ErrLockNotHold,
		},
		{
			name: "lock hold by others",
			before: func(t *testing.T) {
				// 模拟别人持有锁
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "unlock_key2", "value2", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.GetDel(ctx, "unlock_key2").Result()
				require.NoError(t, err)
				// 没释放锁，意味着不变
				assert.Equal(t, "value2", res)
			},
			lock: &distlock.Lock{
				Key:    "unlock_key2",
				Value:  "123",
				Client: rdb,
			},
			wantErr: distlock.ErrLockNotHold,
		},
		{
			name: "unlock",
			before: func(t *testing.T) {
				// 模拟你自己加的锁
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "unlock_key3", "123", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				// 锁被释放，key不存在
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Exists(ctx, "unlock_key3").Result()
				require.NoError(t, err)
				assert.Equal(t, int64(0), res)
			},
			lock: &distlock.Lock{
				Key:    "unlock_key3",
				Value:  "123",
				Client: rdb,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			err := tc.lock.Unlock(ctx)
			assert.Equal(t, tc.wantErr, err)
			tc.after(t)
		})
	}
}
