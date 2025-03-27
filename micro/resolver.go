package micro

import (
	"context"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"imola/micro/registry"
	"time"
)

type GrpcResolverBuilder struct {
	r       registry.Registry
	timeout time.Duration
}

func NewRegistryBuilder(r registry.Registry, timeout time.Duration) (*GrpcResolverBuilder, error) {
	return &GrpcResolverBuilder{
		r:       r,
		timeout: timeout,
	}, nil
}

func (g GrpcResolverBuilder) Build(target resolver.Target,
	cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &grpcResolver{
		cc:      cc,
		r:       g.r,
		target:  target,
		timeout: g.timeout,
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

func (g *grpcResolver) ResolveNow(options resolver.ResolveNowOptions) {
	g.resolve()
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

func (g *grpcResolver) resolve() {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	instances, err := g.r.ListServices(ctx, g.target.Endpoint())
	if err != nil {
		g.cc.ReportError(err)
		return
	}
	address := make([]resolver.Address, 0, len(instances))
	for _, si := range instances {
		address = append(address, resolver.Address{
			Addr:       si.Address,
			Attributes: attributes.New("weight", si.Weight).WithValue("group", si.Group),
		})
	}
	err = g.cc.UpdateState(resolver.State{Addresses: address})
	if err != nil {
		g.cc.ReportError(err)
		return
	}
}

func (g grpcResolver) Close() {
	close(g.close)
}
