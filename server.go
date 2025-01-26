package imola

import (
	"net"
	"net/http"
)

type HandleFunc func(ctx Context)

type Server interface {
	http.Handler
	Start(add string) error

	// AddRoute 路由注册, method是http方法、path是路由、handleFunc是业务逻辑
	addRoute(method string, path string, handleFunc HandleFunc)

	// AddRouteVarFuncs 没有必要去提供，多个函数中断执行、方法优先级etc问题
	// Deprecated
	AddRouteVarFuncs(method string, path string, handleFuncs ...HandleFunc)
}

var _ Server = &HTTPServer{}

type HTTPServer struct {
	*router
}

func NewHTTPServer() *HTTPServer {
	return &HTTPServer{
		router: newRouter(),
	}
}

// AddRouteVarFuncs Deprecated
func (h *HTTPServer) AddRouteVarFuncs(method string, path string, handleFunc ...HandleFunc) {
	panic("implement me")
}

func (h *HTTPServer) addRoute(method string, path string, handleFunc HandleFunc) {
	panic("implement me")
}

// GET 请求，在HTTPServer中实现
func (h *HTTPServer) GET(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodGet, path, handleFunc)
}

// POST 请求，在HTTPServer中实现
func (h *HTTPServer) POST(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodPost, path, handleFunc)
}

func (h *HTTPServer) OPTIONS(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodPost, path, handleFunc)
}

// ServeHTTP HTTPServer 处理请求入口
func (h *HTTPServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := &Context{
		Req:  request,
		Resp: writer,
	}
	h.serve(ctx)
}

// serve 查找路由，执行实际的业务逻辑
func (h *HTTPServer) serve(ctx *Context) {
	// TODO
}

func (h *HTTPServer) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	// 与直接调用http.ListenAndServe(":8081", h)相比，使用HTTPServer可以注册callback function
	return http.Serve(listener, h)
}
