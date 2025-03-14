package distlock

import (
	"context"
	_ "embed"
	"errors"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	ErrFailedToPreemptLock = errors.New("redis-lock: 抢锁失败")
	ErrLockNotHold         = errors.New("redis-lock: 你没有持有锁")
	//go:embed lua/unlock.lua
	LuaUnlock string
)

// Client 就是对redis.Cmdable的二次封装
type Client struct {
	client redis.Cmdable
}

func NewClient(client redis.Cmdable) *Client {
	return &Client{
		client: client,
	}
}

func (c *Client) TryLock(ctx context.Context, key string,
	expiration time.Duration) (*Lock, error) {
	val := uuid.New().String()
	ok, err := c.client.SetNX(ctx, key, val, expiration).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		// 代表的是别人抢到了锁
		return nil, ErrFailedToPreemptLock
	}
	return &Lock{
		Client: c.client,
		Key:    key,
		Value:  val,
	}, nil
}

type Lock struct {
	Client redis.Cmdable
	Key    string
	Value  string
}

func (l *Lock) Unlock(ctx context.Context) error {
	res, err := l.Client.Eval(ctx, LuaUnlock, []string{l.Key}, l.Value).Int64()
	if err != nil {
		return err
	}
	if res != 1 {
		return ErrLockNotHold
	}
	return nil
}
