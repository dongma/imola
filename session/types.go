package session

import (
	"context"
	"net/http"
)

// Session 基本操作，获取、设置session信息
type Session interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, val string) error
	ID() string
}

// Store store管理session,对session提供持久化操作
type Store interface {
	// Generate 生成一个session
	Generate(ctx context.Context, id string) (Session, error)
	// Refresh 刷新同一个session id，使session不失效
	Refresh(ctx context.Context, id string) error
	Remove(ctx context.Context, id string) error
	// Get 获取session信息
	Get(ctx context.Context, id string) (Session, error)
}

// Propagator 抽象层，允许将session id存储在不同地方
type Propagator interface {
	// Inject 将session id注入到http响应里面
	Inject(id string, writer http.ResponseWriter) error
	// Extract 将session id从http请求中解析出来
	Extract(req *http.Request) (string, error)
	// Remove 将session id从http.ResponseWriter中删除
	Remove(writer http.ResponseWriter) error
}
