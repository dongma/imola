package main

import (
	"imola/kernel"
	"imola/kernel/middleware"
	"net/http"
)

func main() {
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
