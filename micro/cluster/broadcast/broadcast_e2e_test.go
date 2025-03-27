package broadcast

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"imola/micro"
	"imola/micro/proto/gen"
	"imola/micro/registry/etcd"
	"testing"
	"time"
)

func TestUseBroadcast(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)
	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)

	var eg errgroup.Group
	var servers []*UserServiceServer
	for i := 0; i < 3; i++ {
		server, err := micro.NewServer("user-service", micro.ServerWithRegistry(r))
		require.NoError(t, err)
		us := &UserServiceServer{
			idx: i,
		}
		servers = append(servers, us)
		gen.RegisterUserServiceServer(server, us)
		// 启动8081，8082和8083 3个服务
		eg.Go(func() error {
			// 在这里调用Start方法，相当于us已完全准备好了
			return server.Start(fmt.Sprintf(":808%d", i+1))
		})
		defer func() {
			_ = server.Close()
		}()
	}
	err = eg.Wait()
	t.Log(err)
	time.Sleep(6 * time.Second)

	client, err := micro.NewClient(micro.ClientInsecure(),
		micro.ClientWithRegistry(r, time.Second*3))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	require.NoError(t, err)

	ctx = UseBroadcast(ctx)
	bd := NewClusterBuilder("user-service", r, grpc.WithInsecure())
	cc, err := client.Dial(ctx, "user-service", grpc.WithUnaryInterceptor(bd.BuildUnaryInterceptor()))
	require.NoError(t, err)

	uc := gen.NewUserServiceClient(cc)
	resp, err := uc.GetById(ctx, &gen.GetByIdReq{Id: 13})
	require.NoError(t, err)
	t.Log(resp)

	// 做断言，广播的话，所有的server都会收到请求
	for _, us := range servers {
		require.Equal(t, 1, us.cnt)
	}
}

type UserServiceServer struct {
	idx int
	cnt int
	gen.UnimplementedUserServiceServer
}

func (s *UserServiceServer) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	s.cnt++
	return &gen.GetByIdResp{
		User: &gen.User{
			Name: fmt.Sprintf("hello, world %d", s.idx),
		},
	}, nil
}
