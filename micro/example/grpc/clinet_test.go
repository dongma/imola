package grpc

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"imola/micro"
	"imola/micro/proto/gen"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	cc, err := grpc.Dial("registry:///localhost:8081", grpc.WithInsecure(), grpc.WithResolvers(&micro.GrpcResolverBuilder{}))
	require.NoError(t, err)
	client := gen.NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	resp, err := client.GetById(ctx, &gen.GetByIdReq{Id: 13})
	require.NoError(t, err)
	t.Log(resp)
}
