package kernel

import (
	"log"
	"net/http"
	"strings"
)

// Core represent core struct
type Core struct {
	router      map[string]*Tree    // 二级map结构
	middlewares []ControllerHandler // 从core这边设置的中间件
}

func NewCore() *Core {
	router := map[string]*Tree{}
	router["GET"] = NewTree()
	router["POST"] = NewTree()
	router["PUT"] = NewTree()
	router["DELETE"] = NewTree()
	core := &Core{router: router}
	return core
}

// Use 注册中间件
func (c *Core) Use(middlewares ...ControllerHandler) {
	c.middlewares = append(c.middlewares, middlewares...)
}

// Get http method wrap，匹配Http请求方法并添加路由规则
func (c *Core) Get(url string, handlers ...ControllerHandler) {
	allHandlers := append(c.middlewares, handlers...)
	if err := c.router["GET"].AddRouter(url, allHandlers); err != nil {
		log.Fatal("add router error: ", err)
	}
}

func (c *Core) Post(url string, handlers ...ControllerHandler) {
	allHandlers := append(c.middlewares, handlers...)
	if err := c.router["POST"].AddRouter(url, allHandlers); err != nil {
		log.Fatal("add router error: ", err)
	}
}

func (c *Core) Put(url string, handlers ...ControllerHandler) {
	allHandlers := append(c.middlewares, handlers...)
	if err := c.router["PUT"].AddRouter(url, allHandlers); err != nil {
		log.Fatal("add router error: ", err)
	}
}

func (c *Core) Delete(url string, handlers ...ControllerHandler) {
	allHandlers := append(c.middlewares, handlers...)
	if err := c.router["DELETE"].AddRouter(url, allHandlers); err != nil {
		log.Fatal("add router error: ", err)
	}
}

// Group === http method wrap end
func (c *Core) Group(prefix string) IGroup {
	return NewGroup(c, prefix)
}

// 匹配路由，如果没有匹配到，则返回nil, 从context.routers中分别按method、uri匹配handler
func (c *Core) findRouteByRequest(request *http.Request) []ControllerHandler {
	// url和method全部转换为大写，保证大小写不敏感
	uri := request.URL.Path
	method := request.Method
	upperMethod := strings.ToUpper(method)

	if methodHandlers, ok := c.router[upperMethod]; ok {
		return methodHandlers.FindHandler(uri)
	}
	return nil
}

func (c *Core) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	log.Println("enter Core#ServeHTTP method, start http server on 8888 port")
	// 封装自定义的context
	ctx := NewContext(request, response)
	// 通过request寻找路由，若未匹配到，则直接返回404的Json
	handlers := c.findRouteByRequest(request)
	if handlers == nil {
		ctx.Json(404, "not found")
		return
	}

	// 设置context中的handlers字段
	ctx.SetHandlers(handlers)
	// 调用路由函数，如果返回err代表内部错误，返回500状态码
	if err := ctx.Next(); err != nil {
		ctx.Json(500, "internal server error")
		return
	}
	log.Println("core.router startup, router /foo rest url to FooControllerHandler")
}
