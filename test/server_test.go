package test

import (
	"fmt"
	"imola"
	"testing"
)

func TestServer(t *testing.T) {
	var server = imola.NewHTTPServer()
	// 1、完全委托给http包
	// http.ListenAndServe(":8081", h)
	// 注册http get请求
	server.GET("/user", func(ctx *imola.Context) {
		fmt.Println("do first thing")
	})

	// 2、自己手动来处理，可以注册listener
	server.Start(":8081")
}
