package test

import (
	"imola"
	"testing"
)

func TestServer_AddRouter(t *testing.T) {
	h := imola.NewHTTPServer()

	h.GET("/order/detail", func(ctx *imola.Context) {
		ctx.Resp.Write([]byte("hello, order detail"))
	})

	h.Start(":8081")
}
