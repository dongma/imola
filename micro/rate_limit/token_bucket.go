package rate_limit

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"time"
)

type TokenBucketLimiter struct {
	Tokens  chan struct{}
	CloseCh chan struct{}
}

// NewTokenBucketLimiter interval表示间隔多久产生一个令牌
func NewTokenBucketLimiter(capacity int, interval time.Duration) *TokenBucketLimiter {
	ch := make(chan struct{}, capacity)
	closeCh := make(chan struct{})
	producer := time.NewTicker(interval)
	go func() {
		defer producer.Stop()
		for {
			select {
			case <-producer.C:
				select {
				case ch <- struct{}{}:
				default:
					// 没人取令牌
				}
			case <-closeCh:
				return
			}
		}
	}()
	return &TokenBucketLimiter{
		Tokens:  ch,
		CloseCh: closeCh,
	}
}

// BuildServerInterceptor 构建服务端限流
func (t *TokenBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// 要在这里拿到令牌
		select {
		case <-t.CloseCh:
			// 已经关掉限流了，因而请求直接放过
			//resp, err = handler(ctx, req)
			err = errors.New("缺乏保护，拒绝请求")
		case <-ctx.Done():
			err = ctx.Err()
			return
		case <-t.Tokens:
			resp, err = handler(ctx, req)
			//default:
			//	err = errors.New("达到瓶颈")
		}
		return
	}
}

func (t *TokenBucketLimiter) Close() error {
	close(t.CloseCh)
	return nil
}
