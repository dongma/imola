package grpc

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/sync/errgroup"
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

	var eg errgroup.Group
	for i := 0; i < 3; i++ {
		var group = "g-A"
		if i%2 == 0 {
			group = "g-B"
		}
		server, err := micro.NewServer("user-service", micro.ServerWithRegistry(r),
			micro.ServerWithGroup(group))
		require.NoError(t, err)
		us := &UserServiceServer{group: group}
		gen.RegisterUserServiceServer(server, us)
		// 启动8081，8082和8083 3个服务
		eg.Go(func() error {
			// 在这里调用Start方法，相当于us已完全准备好了
			return server.Start(fmt.Sprintf(":808%d", i+1))
		})
	}
	err = eg.Wait()
	t.Log(err)
}

type UserServiceServer struct {
	group string
	gen.UnimplementedUserServiceServer
}

func (s UserServiceServer) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	fmt.Println(s.group)
	return &gen.GetByIdResp{
		User: &gen.User{
			Name: "hello, world",
		},
	}, nil
}
