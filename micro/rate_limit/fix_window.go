package rate_limit

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"sync/atomic"
	"time"
)

type FixWindowLimiter struct {
	// 窗口的起始时间
	timestamp int64
	// 窗口大小
	interval int64
	// 在这个窗口内，允许通过的最大请求数量
	rate int64
	cnt  int64
}

func NewFixWindowLimiter(interval time.Duration, rate int64) *FixWindowLimiter {
	return &FixWindowLimiter{
		interval:  int64(interval),
		timestamp: time.Now().UnixNano(),
		rate:      rate,
	}
}

// BuildServerInterceptor 构建服务端限流
func (t *FixWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// 考虑t.cnt重置的问题
		current := time.Now().UnixNano()
		timestamp := atomic.LoadInt64(&t.timestamp)
		cnt := atomic.LoadInt64(&t.cnt)
		if timestamp+t.interval < current {
			// 这意味着这是一个新窗口，重置窗口
			if atomic.CompareAndSwapInt64(&t.timestamp, timestamp, current) {
				atomic.CompareAndSwapInt64(&t.cnt, cnt, 0)
			}
		}

		cnt = atomic.AddInt64(&t.cnt, 1)
		if cnt > t.rate {
			err = errors.New("触发瓶颈了")
			atomic.AddInt64(&t.cnt, -1)
			return
		}
		resp, err = handler(ctx, req)
		return
	}
}
