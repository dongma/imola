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

	server.GET("/values/:id", func(ctx *imola.Context) {
		id, err := ctx.PathValue("id").AsInt64()
		if err != nil {
			ctx.Resp.WriteHeader(400)
			ctx.Resp.Write([]byte("id输入不对"))
			return
		}
		ctx.Resp.Write([]byte(fmt.Sprintf("hello, %d", id)))
	})

	type User struct {
		Name string `json:"name"`
	}

	server.GET("/user/123", func(ctx *imola.Context) {
		ctx.RespJSON(202, User{
			Name: "Tom",
		})
	})
	// 2、自己手动来处理，可以注册listener
	server.Start(":8081")
}
