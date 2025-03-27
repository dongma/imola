package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math"
	"sync"
)

type WeightBalancer struct {
	connections []*weightConn
	mutex       sync.Mutex
}

func (w *WeightBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(w.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	var totalWeight uint32
	var res *weightConn
	for _, conn := range w.connections {
		conn.mutex.Lock()
		totalWeight = totalWeight + conn.efficientWeight
		conn.currentWeight = conn.currentWeight + conn.efficientWeight
		if res == nil || res.currentWeight < conn.currentWeight {
			res = conn
		}
		conn.mutex.Unlock()
	}
	res.mutex.Lock()
	res.currentWeight = res.currentWeight - totalWeight
	res.mutex.Unlock()
	return balancer.PickResult{
		SubConn: res.c,
		Done: func(info balancer.DoneInfo) {
			res.mutex.Lock()
			if info.Err != nil && res.efficientWeight == 0 {
				return
			}
			if info.Err == nil && res.efficientWeight == math.MaxUint32 {
				return
			}
			if info.Err != nil {
				res.efficientWeight--
			} else {
				res.efficientWeight++
			}
			res.mutex.Unlock()
		},
	}, nil
}

type WeightBalancerBuilder struct {
}

func (w *WeightBalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	cs := make([]*weightConn, 0, len(info.ReadySCs))
	for sub, subInfo := range info.ReadySCs {
		weight := subInfo.Address.Attributes.Value("weight").(uint32)
		//weight, err := strconv.ParseUint(weightStr, 10, 64)
		//if err != nil {
		//	panic(err)
		//}
		cs = append(cs, &weightConn{
			c:               sub,
			weight:          uint32(weight),
			currentWeight:   uint32(weight),
			efficientWeight: uint32(weight),
		})
	}
	return &WeightBalancer{
		connections: cs,
	}
}

type weightConn struct {
	mutex sync.Mutex
	c     balancer.SubConn
	// 权重
	weight uint32
	// 当前权重
	currentWeight uint32
	// 有效权重
	efficientWeight uint32
}
