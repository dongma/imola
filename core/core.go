package core

import "net/http"

// 框架的核心结构
type Core struct {
}

// 初始化框架核心结构
func NewCore() *Core {
	return &Core{}
}

// 框架核心结构实现Handler接口
func (c *Core) ServeHttp(response http.ResponseWriter, request *http.Request) {
	// todo
}
