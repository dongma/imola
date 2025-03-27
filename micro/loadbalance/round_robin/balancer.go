package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"imola/micro/loadbalance"
	"sync/atomic"
)

type Balancer struct {
	index       int32
	connections []subConn
	length      int32
	Filter      loadbalance.Filter
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	candidates := make([]subConn, 0, len(b.connections))
	for _, conn := range b.connections {
		if b.Filter != nil && !b.Filter(info, conn.addr) {
			continue
		}
		candidates = append(candidates, conn)
	}
	if len(candidates) == 0 {
		// 也可以考虑在筛选完成后，没有任何符合条件的节点时，返回默认节点
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	idx := atomic.AddInt32(&b.index, 1)
	c := candidates[int(idx)%len(candidates)]
	return balancer.PickResult{
		SubConn: c.c,
		Done: func(info balancer.DoneInfo) {
		},
	}, nil
}

type Builder struct {
	filter loadbalance.Filter
}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]subConn, 0, len(info.ReadySCs))
	for c, ci := range info.ReadySCs {
		connections = append(connections, subConn{
			c:    c,
			addr: ci.Address,
		})
	}
	return &Balancer{
		connections: connections,
		index:       -1,
		length:      int32(len(info.ReadySCs)),
		filter:      b.filter,
	}
}

type subConn struct {
	c    balancer.SubConn
	addr resolver.Address
}
