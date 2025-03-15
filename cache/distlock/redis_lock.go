package distlock

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"time"
)

var (
	ErrFailedToPreemptLock = errors.New("redis-lock: 抢锁失败")
	ErrLockNotHold         = errors.New("redis-lock: 你没有持有锁")
	//go:embed lua/unlock.lua
	LuaUnlock string
	//go:embed lua/refresh.lua
	LuaRefresh string
	//go:embed lua/lock.lua
	LuaLock string
)

// Client 就是对redis.Cmdable的二次封装
type Client struct {
	client redis.Cmdable
	g      singleflight.Group
}

func NewClient(client redis.Cmdable) *Client {
	return &Client{
		client: client,
	}
}

func (c *Client) SingleFlightLock(ctx context.Context, key string,
	expiration time.Duration,
	timeout time.Duration,
	retry RetryStrategy) (*Lock, error) {
	for {
		flag := false
		resCh := c.g.DoChan(key, func() (interface{}, error) {
			flag = true
			return c.Lock(ctx, key, expiration, timeout, retry)
		})
		select {
		case res := <-resCh:
			if flag {
				c.g.Forget(key)
				if res.Err != nil {
					return nil, res.Err
				}
				return res.Val.(*Lock), nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (c *Client) Lock(ctx context.Context, key string,
	expiration time.Duration,
	timeout time.Duration,
	retry RetryStrategy) (*Lock, error) {
	var timer *time.Timer
	val := uuid.New().String()
	for {
		// 在这里重试
		lctx, cancel := context.WithTimeout(ctx, timeout)
		res, err := c.client.Eval(lctx, LuaLock, []string{key}, val, expiration.Seconds()).Result()
		cancel()
		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		if res == "OK" {
			return &Lock{
				Client:     c.client,
				Key:        key,
				Value:      val,
				Expiration: expiration,
				UnlockChan: make(chan struct{}, 1),
			}, nil
		}

		interval, ok := retry.Next()
		if !ok {
			return nil, fmt.Errorf("redis-lock: 超出重试限制, %w", ErrFailedToPreemptLock)
		}
		if timer == nil {
			timer = time.NewTimer(interval)
		} else {
			timer.Reset(interval)
		}

		select {
		case <-timer.C:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
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
		Client:     c.client,
		Key:        key,
		Value:      val,
		Expiration: expiration,
		UnlockChan: make(chan struct{}, 1),
	}, nil
}

type Lock struct {
	Client     redis.Cmdable
	Key        string
	Value      string
	Expiration time.Duration
	UnlockChan chan struct{}
}

func (l *Lock) AutoRefresh(interval time.Duration, timeout time.Duration) error {
	timeoutChan := make(chan struct{}, 1)

	// 间隔多久续约一次
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			// 定时刷新
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			err := l.Refresh(ctx)
			// 出现err后，也要向errChan发消息，同时关闭channel
			if err == context.DeadlineExceeded {
				timeoutChan <- struct{}{}
				continue
			}
			if err != nil {
				return err
			}
		case <-timeoutChan:
			// 定时刷新
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			err := l.Refresh(ctx)
			// 出现err后，也要向errChan发消息，同时关闭channel
			if err == context.DeadlineExceeded {
				timeoutChan <- struct{}{}
				continue
			}
			if err != nil {
				return err
			}
		case <-l.UnlockChan:
			return nil
		}
	}
}

func (l *Lock) Refresh(ctx context.Context) error {
	res, err := l.Client.Eval(ctx, LuaRefresh, []string{l.Key}, l.Value, l.Expiration.Seconds()).Int64()
	defer func() {
		close(l.UnlockChan)
	}()
	if err != nil {
		return err
	}
	if res != 1 {
		return ErrLockNotHold
	}
	return nil
}

func (l *Lock) Unlock(ctx context.Context) error {
	res, err := l.Client.Eval(ctx, LuaUnlock, []string{l.Key}, l.Value).Int64()
	defer func() {
		select {
		case l.UnlockChan <- struct{}{}:
		default:
			// 说明没有人调用 AutoRefresh
		}
	}()
	if err != nil {
		return err
	}
	if res != 1 {
		return ErrLockNotHold
	}
	return nil
}
