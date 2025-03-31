package test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"imola/micro/proto/gen"
	"imola/micro/rate_limit"
	"testing"
	"time"
)

func TestTokenBucketLimiter_BuildServerInterceptor(t *testing.T) {
	testCases := []struct {
		name     string
		b        func() *rate_limit.TokenBucketLimiter
		ctx      context.Context
		handler  func(ctx context.Context, req interface{}) (interface{}, error)
		wantErr  error
		wantResp any
	}{
		{
			name: "closed",
			b: func() *rate_limit.TokenBucketLimiter {
				closeCh := make(chan struct{})
				close(closeCh)
				return &rate_limit.TokenBucketLimiter{
					Tokens:  make(chan struct{}),
					CloseCh: closeCh,
				}
			},
			ctx:     context.Background(),
			wantErr: errors.New("缺乏保护，拒绝请求"),
		},
		{
			name: "context canceled",
			b: func() *rate_limit.TokenBucketLimiter {
				return &rate_limit.TokenBucketLimiter{
					Tokens:  make(chan struct{}),
					CloseCh: make(chan struct{}),
				}
			},
			// 传进去的context是一个已经被关掉的context
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
		{
			name: "get tokens",
			b: func() *rate_limit.TokenBucketLimiter {
				ch := make(chan struct{}, 1)
				ch <- struct{}{}
				return &rate_limit.TokenBucketLimiter{
					Tokens:  ch,
					CloseCh: make(chan struct{}),
				}
			},
			ctx: context.Background(),
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return &gen.GetByIdResp{}, errors.New("mock error")
			},
			wantErr:  errors.New("mock error"),
			wantResp: &gen.GetByIdResp{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interceptor := tc.b().BuildServerInterceptor()
			resp, err := interceptor(tc.ctx, &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, tc.handler)
			assert.Equal(t, tc.wantErr, err)
			if err == nil {
				return
			}
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func TestTokenBucketLimiter_Tokens(t *testing.T) {
	limiter := rate_limit.NewTokenBucketLimiter(10, 2*time.Second)
	defer limiter.Close()
	interceptor := limiter.BuildServerInterceptor()
	cnt := 0
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		cnt++
		return &gen.GetByIdResp{}, nil
	}
	resp, err := interceptor(context.Background(), &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	require.NoError(t, err)
	assert.Equal(t, &gen.GetByIdResp{}, resp)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()
	// 触发限流
	resp, err = interceptor(ctx, &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	assert.Equal(t, context.DeadlineExceeded, err)
	require.Nil(t, resp)
}
