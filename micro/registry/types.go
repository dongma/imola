package registry

import (
	"context"
	"io"
)

// Registry 定义注册中心接口
type Registry interface {
	Register(ctx context.Context, si ServiceInstance) error
	Unregister(ctx context.Context, si ServiceInstance) error
	ListServices(ctx context.Context, serviceName string) ([]ServiceInstance, error)
	Subscribe(serviceName string) (<-chan Event, error)

	io.Closer
}

type ServiceInstance struct {
	Name string
	// Address 就是最关键的，定位信息
	Address string
}

type Event struct {
}
