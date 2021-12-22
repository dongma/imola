package kernel

import (
	"strings"
	"log"
	"net/http"
)

// Core represent core struct
type Core struct {
	router map[string]map[string]ControllerHandler 	// 二级map结构
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



func (c *Core) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	log.Println("enter Core#ServeHTTP method, start http server on 8888 port")
	ctx := NewContext(request, response)

	router := c.router["foo"]
	if router == nil {
		return
	} 
	log.Println("core.router startup, router /foo rest url to FooControllerHandler")
	router(ctx)
}
