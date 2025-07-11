package broadcast

import (
	"context"
	"fmt"
	"github.com/dongma/imola/micro/registry"
	"google.golang.org/grpc"
	"reflect"
	"sync"
)

type ClusterBuilder struct {
	registry registry.Registry
	service  string
	dialOpts []grpc.DialOption
}

func NewClusterBuilder(service string, r registry.Registry, dOpts ...grpc.DialOption) *ClusterBuilder {
	return &ClusterBuilder{
		service:  service,
		registry: r,
		dialOpts: dOpts,
	}
}

func (c ClusterBuilder) BuildUnaryInterceptor() grpc.UnaryClientInterceptor {
	// method: users.UserService/GetById
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ok, ch := isBroadCast(ctx)
		if !ok {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		defer func() {
			close(ch)
		}()

		instances, err := c.registry.ListServices(ctx, c.service)
		if err != nil {
			return err
		}

		var wg sync.WaitGroup
		typ := reflect.TypeOf(reply).Elem()
		wg.Add(len(instances))
		for _, ins := range instances {
			addr := ins.Address
			go func() {
				insCC, er := grpc.Dial(addr, c.dialOpts...)
				if er != nil {
					ch <- Resp{Err: er}
					wg.Done()
					return
				}
				newReply := reflect.New(typ).Interface()
				// 对每一个节点 发起调用
				err = invoker(ctx, method, req, newReply, insCC, opts...)
				// 如果没有人接收，就会堵住
				select {
				case <-ctx.Done():
					err = fmt.Errorf("响应没有人接收, %w", ctx.Err())
				case ch <- Resp{Err: err, Reply: newReply}:
				}
				wg.Done()
			}()
		}
		wg.Wait()
		// 全部调用完毕
		return nil
	}
}

func UseBroadcast(ctx context.Context) (context.Context, <-chan Resp) {
	ch := make(chan Resp)
	return context.WithValue(ctx, broadcastKey{}, ch), ch
}

type broadcastKey struct {
}

func isBroadCast(ctx context.Context) (bool, chan Resp) {
	val, ok := ctx.Value(broadcastKey{}).(chan Resp)
	return ok, val
}

type Resp struct {
	Err   error
	Reply any
}
