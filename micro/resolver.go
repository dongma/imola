package micro

import (
	"context"
	"google.golang.org/grpc/resolver"
	"imola/micro/registry"
	"time"
)

type GrpcResolverBuilder struct {
	r       registry.Registry
	timeout time.Duration
}

func NewRegistryBuilder(r registry.Registry) (*GrpcResolverBuilder, error) {
	return &GrpcResolverBuilder{r: r}, nil
}

func (g GrpcResolverBuilder) Build(target resolver.Target,
	cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &grpcResolver{
		cc:     cc,
		r:      g.r,
		target: target,
	}
	r.resolve()
	go r.watch()
	return r, nil
}

func (g GrpcResolverBuilder) Scheme() string {
	return "registry"
}

type grpcResolver struct {
	target  resolver.Target
	r       registry.Registry
	cc      resolver.ClientConn
	timeout time.Duration
	close   chan struct{}
}

func (g grpcResolver) ResolveNow(options resolver.ResolveNowOptions) {
	g.resolve()
}

func (g *grpcResolver) resolve() {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	instances, err := g.r.ListServices(ctx, g.target.Endpoint())
	if err != nil {
		g.cc.ReportError(err)
		return
	}
	address := make([]resolver.Address, len(instances))
	for _, si := range instances {
		address = append(address, resolver.Address{Addr: si.Address})
	}
	err = g.cc.UpdateState(resolver.State{Addresses: address})
	if err != nil {
		g.cc.ReportError(err)
		return
	}
}

func (g grpcResolver) watch() {
	events, err := g.r.Subscribe(g.target.Endpoint())
	if err != nil {
		g.cc.ReportError(err)
		return
	}
	for {
		select {
		case <-events:
			g.resolve()
		case <-g.close:
			return
		}
	}
}

func (g grpcResolver) Close() {
	close(g.close)
}
