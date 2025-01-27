package imola

import (
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	h := NewHTTPServer()

	h.addRoute(http.MethodGet, "/order/detail", func(ctx *Context) {
		ctx.Resp.Write([]byte("hello, order detail"))
	})

	h.Start(":8081")
}
