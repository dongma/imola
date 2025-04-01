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

func TestFixWindowLimiter_BuildServerInterceptor(t *testing.T) {
	// 三秒钟只能有一个请求
	interceptor := rate_limit.NewFixWindowLimiter(time.Second*3, 1).BuildServerInterceptor()
	cnt := 0
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		cnt++
		return &gen.GetByIdReq{}, nil
	}
	// 第一个肯定能通过
	resp, err := interceptor(context.Background(), &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	require.NoError(t, err)
	assert.Equal(t, &gen.GetByIdReq{}, resp)

	// 第二个肯定触发了瓶颈
	resp, err = interceptor(context.Background(), &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	require.Equal(t, errors.New("触发瓶颈了"), err)
	assert.Nil(t, resp)

	// 睡一个三秒，确保窗口新建了
	time.Sleep(time.Second * 3)
	resp, err = interceptor(context.Background(), &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	require.NoError(t, err)
	assert.Equal(t, &gen.GetByIdReq{}, resp)
}
