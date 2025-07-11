package test

import (
	"github.com/dongma/imola/web"
	"testing"
)

func TestServer_AddRouter(t *testing.T) {
	h := web.NewHTTPServer()

	h.GET("/order/detail", func(ctx *web.Context) {
		ctx.Resp.Write([]byte("hello, order detail"))
	})

	h.Start(":8081")
}
