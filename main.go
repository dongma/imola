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
	// core使用use注册中间件
	core.Use(middleware.Test1(), middleware.Test2())
	registerRouter(core)
	subjectApi := core.Group("/subject")
	subjectApi.Use(middleware.Test2())
	server := &http.Server{
		Handler: core,
		Addr:    "localhost:8888",
	}
	server.ListenAndServe()
}
