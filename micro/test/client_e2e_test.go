package test

import (
	"context"
	"errors"
	"github.com/dongma/imola/micro/proto/gen"
	"github.com/dongma/imola/micro/rpc"
	"github.com/dongma/imola/micro/rpc/serialize/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestInitServiceProto(t *testing.T) {
	server := rpc.NewServer()
	service := &rpc.UserServiceServer{}

	server.RegisterService(service)
	server.RegisterSerializer(&proto.Serializer{})
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)
	usClient := &rpc.UserService{}

	client, err := rpc.NewClient(":8081", rpc.ClientWithSerializer(&proto.Serializer{}))
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		mock     func()
		wantErr  error
		wantResp *rpc.GetByIdResp
	}{
		{
			name: "no error",
			mock: func() {
				service.Err = nil
				service.Msg = "hello, world"
			},
			wantResp: &rpc.GetByIdResp{
				Msg: "hello, world",
			},
		},
		{
			name: "error",
			mock: func() {
				service.Msg = ""
				service.Err = errors.New("mock error")
			},
			wantResp: &rpc.GetByIdResp{},
			wantErr:  errors.New("mock error"),
		},
		{
			name: "both",
			mock: func() {
				service.Msg = "hello, world"
				service.Err = errors.New("mock error")
			},
			wantResp: &rpc.GetByIdResp{
				Msg: "hello, world",
			},
			wantErr: errors.New("mock error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			resp, er := usClient.GetByIdProto(context.Background(), &gen.GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, er)
			if resp != nil && resp.User != nil {
				assert.Equal(t, tc.wantResp.Msg, resp.User.Name)
			}
		})
		tc.mock()
	}

}

func TestOneway(t *testing.T) {
	server := rpc.NewServer()
	service := &rpc.UserServiceServer{}
	server.RegisterService(service)
	go func() {
		err := server.Start("tcp", ":8082")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)
	usClient := &rpc.UserService{}

	client, err := rpc.NewClient(":8082")
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		mock     func()
		wantErr  error
		wantResp *rpc.GetByIdResp
	}{
		{
			name: "oneway",
			mock: func() {
				service.Err = errors.New("mock error")
				service.Msg = "hello, world"
			},
			wantResp: &rpc.GetByIdResp{},
			wantErr:  errors.New("micro: 这是一个oneway调用，你不应该处理任何结果"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			ctx := rpc.CtxWithOneway(context.Background())
			resp, er := usClient.GetById(ctx, &rpc.GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, er)
			assert.Equal(t, tc.wantResp, resp)
			time.Sleep(time.Second * 2)
			assert.Equal(t, "hello, world", service.Msg)
		})
	}

}

func TestInitClientProxy(t *testing.T) {
	server := rpc.NewServer()
	service := &rpc.UserServiceServer{}
	server.RegisterService(service)
	go func() {
		err := server.Start("tcp", ":8083")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)
	usClient := &rpc.UserService{}

	client, err := rpc.NewClient(":8083")
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		mock     func()
		wantErr  error
		wantResp *rpc.GetByIdResp
	}{
		{
			name: "no error",
			mock: func() {
				service.Err = nil
				service.Msg = "hello, world"
			},
			wantResp: &rpc.GetByIdResp{
				Msg: "hello, world",
			},
		},
		{
			name: "error",
			mock: func() {
				service.Msg = ""
				service.Err = errors.New("mock error")
			},
			wantResp: &rpc.GetByIdResp{},
			wantErr:  errors.New("mock error"),
		},
		{
			name: "both",
			mock: func() {
				service.Msg = "hello, world"
				service.Err = errors.New("mock error")
			},
			wantResp: &rpc.GetByIdResp{
				Msg: "hello, world",
			},
			wantErr: errors.New("mock error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			// 在用CtxWithOneway的情况下，不会有任何返回
			// resp, er := usClient.GetById(rpc.CtxWithOneway(context.Background()), &rpc.GetByIdReq{Id: 123})

			resp, er := usClient.GetById(context.Background(), &rpc.GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, er)
			assert.Equal(t, tc.wantResp, resp)
		})
	}

}

func TestTimeout(t *testing.T) {
	server := rpc.NewServer()
	service := &rpc.UserServiceServerTimeout{T: t}
	server.RegisterService(service)
	go func() {
		err := server.Start("tcp", ":8084")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)

	usClient := &rpc.UserService{}
	client, err := rpc.NewClient(":8084")
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		mock     func() context.Context
		wantErr  error
		wantResp *rpc.GetByIdResp
	}{
		{
			name: "service",
			mock: func() context.Context {
				service.Err = errors.New("mock error")
				service.Msg = "hello, world"
				// 测试场景，服务睡眠2s，但是超时设置了1s，所以客户端预期拿到超时响应
				service.Sleep = time.Second * 2
				ctx, _ := context.WithTimeout(context.Background(), time.Second)
				return ctx
			},
			wantResp: &rpc.GetByIdResp{},
			wantErr:  context.DeadlineExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, er := usClient.GetById(tc.mock(), &rpc.GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, er)
			assert.Equal(t, tc.wantResp, resp)
		})
	}

}
