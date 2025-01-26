package test

import (
	"fmt"
	"imola"
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	var h = &imola.HTTPServer{}
	// 1、完全委托给http包
	// http.ListenAndServe(":8081", h)

	h.AddRoute(http.MethodGet, "/user", func(ctx imola.Context) {
		fmt.Println("do first thing")
		fmt.Println("do second thing")
	})

	h.AddRouteVarFuncs(http.MethodGet, "/user", func(ctx imola.Context) {
		fmt.Println("do first thing")
	}, func(ctx imola.Context) {
		fmt.Println("do second thing")
	})

	// 注册http get请求
	h.GET("/user", func(ctx imola.Context) {
		fmt.Println("do first thing")
	})

	// 2、自己手动来处理，可以注册listener
	h.Start(":8081")
}
