package kernel

import (
	"log"
	"net/http"
)

// Core represent core struct
type Core struct {
	router map[string]ControllerHandler
}

func NewCore() *Core {
	return &Core{router: map[string]ControllerHandler{}}
}

func (c *Core) Get(url string, handler ControllerHandler) {
	c.router[url] = handler
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
