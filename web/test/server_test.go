package test

import (
	"fmt"
	"github.com/dongma/imola/web"
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	var server = web.NewHTTPServer()
	// 1、完全委托给http包
	// http.ListenAndServe(":8081", h)
	// 注册http get请求
	server.GET("/user", func(ctx *web.Context) {
		fmt.Println("do first thing")
	})

	server.GET("/values/:id", func(ctx *web.Context) {
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

	server.GET("/user/123", func(ctx *web.Context) {
		ctx.RespJSON(202, User{
			Name: "Tom",
		})
	})
	// 2、自己手动来处理，可以注册listener
	server.Start(":8081")
}

func TestHttpServer_serverHTTP(t *testing.T) {
	server := web.NewHTTPServer()
	middles := []web.Middleware{
		func(next web.HandleFunc) web.HandleFunc {
			return func(ctx *web.Context) {
				fmt.Println("第一个before...")
				next(ctx)
				fmt.Println("第一个after....")
			}
		},
		func(next web.HandleFunc) web.HandleFunc {
			return func(ctx *web.Context) {
				fmt.Println("第二个before...")
				next(ctx)
				fmt.Println("第二个after....")
			}
		},
		func(next web.HandleFunc) web.HandleFunc {
			return func(ctx *web.Context) {
				fmt.Println("第三个中断...")
			}
		},
		func(next web.HandleFunc) web.HandleFunc {
			return func(ctx *web.Context) {
				fmt.Println("第四个看不到...")
			}
		},
	}
	server.SetMiddlewares(middles)
	server.ServeHTTP(nil, &http.Request{})
}
