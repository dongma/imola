package recover

import (
	"fmt"
	"imola"
	"testing"
)

func TestMiddlewareBuilder_Build(t *testing.T) {
	builder := MiddlewareBuilder{
		StatusCode: 500,
		Data:       []byte("你panic了"),
		Log: func(ctx *imola.Context) {
			fmt.Printf("panic路径: %s", ctx.Req.URL.String())
		},
	}

	server := imola.NewHTTPServer(imola.ServerWithMiddleware(builder.Build()))
	server.GET("/user", func(ctx *imola.Context) {
		panic("发生panic了")
	})
	server.Start(":8081")
}
