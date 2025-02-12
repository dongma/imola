package imola

import (
	"fmt"
	"net"
	"net/http"
)

type HandleFunc func(ctx *Context)

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

type HTTPServerOption func(server *HTTPServer)

type HTTPServer struct {
	router
	mids []Middleware

	log func(msg string, args ...any)

	tplEngine TemplateEngine
}

func NewHTTPServer(opts ...HTTPServerOption) *HTTPServer {
	res := &HTTPServer{
		router: newRouter(),
		log: func(msg string, args ...any) {
			fmt.Printf(msg, args...)
		},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func ServerWithMiddleware(mdls ...Middleware) HTTPServerOption {
	return func(server *HTTPServer) {
		server.mids = mdls
	}
}

func ServerWithTemplateEngine(tplEngine TemplateEngine) HTTPServerOption {
	return func(server *HTTPServer) {
		server.tplEngine = tplEngine
	}
}

// SetMiddlewares Deprecated，方法不优雅，代码太硬没有设计
func (h *HTTPServer) SetMiddlewares(middlewares []Middleware) {
	h.mids = middlewares
}

// AddRouteVarFuncs Deprecated
func (h *HTTPServer) AddRouteVarFuncs(method string, path string, handleFunc ...HandleFunc) {
	panic("implement me")
}

func (h *HTTPServer) addRoute(method string, path string, handleFunc HandleFunc) {
	h.router.addRoute(method, path, handleFunc)
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
		Req:       request,
		Resp:      writer,
		tplEngine: h.tplEngine,
	}
	root := h.serve
	// 然后这里就是调用最后一个不断向前回溯的组装链条，从后往前构造一个链条
	for i := len(h.mids) - 1; i >= 0; i-- {
		root = h.mids[i](root)
	}

	// 这里是最后一个步骤，就是把RespData和RespStatusCode 刷新到响应里面
	var m Middleware = func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			next(ctx)
			h.flashResp(ctx)
		}
	}
	root = m(root)
	root(ctx)
}

func (h *HTTPServer) flashResp(ctx *Context) {
	if ctx.RespStatusCode != 0 {
		ctx.Resp.WriteHeader(ctx.RespStatusCode)
	}
	n, err := ctx.Resp.Write(ctx.RespData)
	if err != nil || n != len(ctx.RespData) {
		h.log("写入响应失败... %v", err)
	}
}

// serve 查找路由，执行实际的业务逻辑
func (h *HTTPServer) serve(ctx *Context) {
	info, ok := h.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok || info.n.handler == nil {
		// 路由没有命中，就是404
		/*ctx.Resp.WriteHeader(404)
		_, _ = ctx.Resp.Write([]byte("not found"))*/
		ctx.RespStatusCode = 404
		ctx.RespData = []byte("not found")
		return
	}
	ctx.PathParams = info.pathParams
	ctx.MatchedRoute = info.n.route
	info.n.handler(ctx)
}

func (h *HTTPServer) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	// 与直接调用http.ListenAndServe(":8081", h)相比，使用HTTPServer可以注册callback function
	return http.Serve(listener, h)
}
