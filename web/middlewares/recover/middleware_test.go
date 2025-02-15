package recover

import (
	"fmt"
	"imola/web"
	"testing"
)

func TestMiddlewareBuilder_Build(t *testing.T) {
	builder := MiddlewareBuilder{
		StatusCode: 500,
		Data:       []byte("你panic了"),
		Log: func(ctx *web.Context) {
			fmt.Printf("panic路径: %s", ctx.Req.URL.String())
		},
	}

	server := web.NewHTTPServer(web.ServerWithMiddleware(builder.Build()))
	server.GET("/user", func(ctx *web.Context) {
		panic("发生panic了")
	})
	server.Start(":8081")
}
