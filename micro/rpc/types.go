package rpc

import (
	"context"
	"imola/micro/rpc/protocol"
)

type Service interface {
	Name() string
}

type Proxy interface {
	Invoke(ctx context.Context, req *protocol.Request) (*protocol.Response, error)
}
