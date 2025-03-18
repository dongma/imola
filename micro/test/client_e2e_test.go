package test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"imola/micro/rpc"
	"testing"
	"time"
)

func TestInitClientProxy(t *testing.T) {
	server := rpc.NewServer()
	server.RegisterService(&rpc.UserServiceServer{})
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)
	usClient := &rpc.UserService{}
	err := rpc.InitClientProxy(":8081", usClient)
	require.NoError(t, err)
	resp, err := usClient.GetById(context.Background(), &rpc.GetByIdReq{Id: 123})
	require.NoError(t, err)
	assert.Equal(t, &rpc.GetByIdResp{Msg: "hello, world"}, resp)
}
