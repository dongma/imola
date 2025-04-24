package accesslog

import (
	"fmt"
	"imola/web"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareBuilder_Build(t *testing.T) {
	builder := MiddlewareBuilder{}
	mdl1 := builder.LogFunc(func(log string) {
		fmt.Println(log)
	}).Build()
	mdl2 := builder.LogFunc(func(log string) {
		fmt.Println("second mdl: " + log)
	}).Build()
	server := web.NewHTTPServer(web.ServerWithMiddleware(mdl1, mdl2))
	server.POST("/a/b/*", func(ctx *web.Context) {
		fmt.Println("hello, it's me")
	})
	req, err := http.NewRequest(http.MethodPost, "/a/b/c", nil)
	req.Host = "localhost"
	if err != nil {
		t.Fatal(err)
	}
	//=== RUN   TestMiddlewareBuilder_Build
	//hello, it's me
	//{"host":"localhost","route":"a/b/*","http_method":"POST","path":"/a/b/c"}
	server.ServeHTTP(httptest.NewRecorder(), req)
}
