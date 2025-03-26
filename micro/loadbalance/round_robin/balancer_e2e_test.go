package round_robin

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"imola/micro"
	"imola/micro/proto/gen"
	"imola/micro/registry/etcd"
	"testing"
	"time"
)

func TestBalancer_e2e_Pick(t *testing.T) {
	go func() {
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
	}()

	time.Sleep(5 * time.Second)
	balancer.Register(base.NewBalancerBuilder("DEMO_ROUND_ROBIN", &Builder{},
		base.Config{HealthCheck: true}))
	cc, err := grpc.Dial("localhost:8081", grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"LoadBalancingPolicy": "DEMO_ROUND_ROBIN"}`))
	require.NoError(t, err)
	client := gen.NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	resp, err := client.GetById(ctx, &gen.GetByIdReq{Id: 13})
	require.NoError(t, err)
	t.Log(resp)
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
