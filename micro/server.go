package micro

import (
	"context"
	"google.golang.org/grpc"
	"imola/micro/registry"
	"net"
	"time"
)

type ServerOption func(server *Server)

type Server struct {
	*grpc.Server
	name            string
	registry        registry.Registry
	registryTimeout time.Duration
	listener        net.Listener
}

func NewServer(name string, opts ...ServerOption) (*Server, error) {
	res := &Server{
		name:            name,
		Server:          grpc.NewServer(),
		registryTimeout: 10 * time.Second,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

// Start 当用户调用这个方法的时候，就是服务已经准备好
func (s *Server) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = listener

	// 有注册中心，要注册了
	if s.registry == nil {
		ctx, cancel := context.WithTimeout(context.Background(), s.registryTimeout)
		defer cancel()
		err = s.registry.Register(ctx, registry.ServiceInstance{
			Name: s.name,
			// 你的定位信息从哪儿来？在容器中时，可以从环境变量中读取
			Address: listener.Addr().String(),
		})
		if err != nil {
			return err
		}
		// 在这里已注册成功
		//defer func() {
		// 忽略或者log一下错误
		//_ = s.registry.Close()
		//}()
	}
	err = s.Serve(listener)
	return err
}

func (s *Server) Close(ctx context.Context) error {
	if s.registry != nil {
		err := s.registry.Close()
		if err != nil {
			return err
		}
	}
	s.GracefulStop()
	return nil
}

func ServerWithRegistry(registry registry.Registry) ServerOption {
	return func(server *Server) {
		server.registry = registry
	}
}
