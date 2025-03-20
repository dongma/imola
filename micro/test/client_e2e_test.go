package test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"imola/micro/rpc"
	"testing"
	"time"
)

func TestInitClientProxy(t *testing.T) {
	server := rpc.NewServer()
	service := &rpc.UserServiceServer{}
	server.RegisterService(service)
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)
	usClient := &rpc.UserService{}
	err := rpc.InitClientProxy(":8081", usClient)
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
			resp, er := usClient.GetById(context.Background(), &rpc.GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, er)
			assert.Equal(t, tc.wantResp, resp)
		})
		tc.mock()
	}

}
