package main

import (
	"github.com/sirupsen/logrus"
	"imola/kernel"
	"imola/kernel/middleware"
	"net/http"
)

func main() {
	logrus.Println("hello, using [go mod] tool to manage dependencies")
	core := kernel.NewCore()
	// core使用use注册中间件，构建go web服务的脚本:'go mod tidy', then execute:'go build'
	//core.Use(middleware.Test1(), middleware.Test2())
	core.Use(middleware.Recovery())
	registerRouter(core)

	// 2.使用api url进行接口测试，curl http://localhost:8888/user/login, "ok, UserLoginController"
	subjectApi := core.Group("/subject")
	subjectApi.Use(middleware.Test2())
	server := &http.Server{
		Handler: core,
		Addr:    "localhost:8888",
	}
	server.ListenAndServe()
}
