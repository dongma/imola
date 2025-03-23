package grpc

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"imola/micro"
	"imola/micro/proto/gen"
	"imola/micro/registry/etcd"
	"testing"
)

func TestServer(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)
	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)
	us := &UserServiceServer{}
	server, err := micro.NewServer("user-service", micro.ServerWithRegistry(r))
	require.NoError(t, err)
	gen.RegisterUserServiceServer(server, us)

	// 在这里调用Start方法，相当于us已完全准备好了
	err = server.Start(":8081")
	t.Log(err)
}

type UserServiceServer struct {
	gen.UnimplementedUserServiceServer
}

func (s UserServiceServer) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	fmt.Println(req)
	return &gen.GetByIdResp{
		User: &gen.User{
			Name: "hello, world",
		},
	}, nil
}
