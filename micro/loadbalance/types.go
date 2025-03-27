package loadbalance

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

// Filter 函数返回值为true时留下来，返回false时过滤掉
type Filter func(info balancer.PickInfo, addr resolver.Address) bool

type GroupFilterBuilder struct {
	Group string
}

func (g GroupFilterBuilder) Build() Filter {
	return func(info balancer.PickInfo, addr resolver.Address) bool {
		target := addr.Attributes.Value("group").(string)
		input := info.Ctx.Value("group").(string)
		return target == input
	}
}
