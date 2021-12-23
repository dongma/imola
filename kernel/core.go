package kernel

import (
	"log"
	"net/http"
	"strings"
)

// Core represent core struct
type Core struct {
	router map[string]map[string]ControllerHandler // 二级map结构
}

func NewCore() *Core {
	// 按请求方法定义二级map，并将其注册到一级map中
	getRouter := map[string]ControllerHandler{}
	postRouter := map[string]ControllerHandler{}
	putRouter := map[string]ControllerHandler{}
	deleteRouter := map[string]ControllerHandler{}

	router := map[string]map[string]ControllerHandler{}
	router["GET"] = getRouter
	router["POST"] = postRouter
	router["PUT"] = putRouter
	router["DELETE"] = deleteRouter
	return &Core{router: router}
}

// http method wrap，匹配Http请求方法并添加路由规则
func (c *Core) Get(url string, handler ControllerHandler) {
	upperUrl := strings.ToUpper(url)
	c.router["GET"][upperUrl] = handler
}

func (c *Core) Post(url string, handler ControllerHandler) {
	upperUrl := strings.ToUpper(url)
	c.router["POST"][upperUrl] = handler
}

func (c *Core) Put(url string, handler ControllerHandler) {
	upperUrl := strings.ToUpper(url)
	c.router["PUT"][upperUrl] = handler
}

func (c *Core) Delete(url string, handler ControllerHandler) {
	upperUrl := strings.ToUpper(url)
	c.router["DELETE"][upperUrl] = handler
}

// === http method wrap end
func (c *Core) Group(prefix string) IGroup {
	return NewGroup(c, prefix)
}

// 匹配路由，如果没有匹配到，则返回nil, 从context.routers中分别按method、uri匹配handler
func (c *Core) findRouteByRequest(request *http.Request) ControllerHandler {
	// url和method全部转换为大写，保证大小写不敏感
	uri := request.URL.Path
	method := request.Method
	upperMethod := strings.ToUpper(method)
	upperUri := strings.ToUpper(uri)

	if methodHandlers, ok := c.router[upperMethod]; ok {
		if handler, ok := methodHandlers[upperUri]; ok {
			return handler
		}
	}
	return nil
}

func (c *Core) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	log.Println("enter Core#ServeHTTP method, start http server on 8888 port")
	ctx := NewContext(request, response)

	// 通过request寻找路由，若未匹配到，则直接返回404的Json
	router := c.findRouteByRequest(request)
	if router == nil {
		ctx.Json(404, "not found")
		return
	}

	if err := router(ctx); err != nil {
		ctx.Json(500, "internal server error")
		return
	}
	log.Println("core.router startup, router /foo rest url to FooControllerHandler")
}
