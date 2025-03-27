package grpc

import (
	"context"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"imola/micro"
	"imola/micro/loadbalance"
	"imola/micro/loadbalance/round_robin"
	"imola/micro/proto/gen"
	"imola/micro/registry/etcd"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)
	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)

	client, err := micro.NewClient(micro.ClientInsecure(),
		micro.ClientWithRegistry(r, time.Second*3),
		micro.ClientWithPickedBuilder("GROUP_ROUND_ROBIN", &round_robin.Builder{
			Filter: loadbalance.GroupFilterBuilder{}.Build(),
		}))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	require.NoError(t, err)

	ctx = context.WithValue(ctx, "group", "g-A")
	cc, err := client.Dial(ctx, "user-service")
	require.NoError(t, err)

	uc := gen.NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		resp, err := uc.GetById(ctx, &gen.GetByIdReq{Id: 13})
		require.NoError(t, err)
		t.Log(resp)
	}
}
