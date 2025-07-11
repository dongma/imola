package test

import (
	"github.com/dongma/imola/web"
	"github.com/dongma/imola/web/session"
	"github.com/dongma/imola/web/session/cookie"
	"github.com/dongma/imola/web/session/memory"
	"github.com/google/uuid"
	"net/http"
	"testing"
	"time"
)

func TestManager(t *testing.T) {
	server := web.NewHTTPServer()
	manager := session.Manager{
		SessionCtxKey: "_sess",
		Store:         memory.NewStore(30 * time.Minute),
		Propagator: cookie.NewPropagator("sessid",
			cookie.WithCookieOption(func(c *http.Cookie) {
				c.HttpOnly = true
			})),
	}

	server.GET("/login", func(ctx *web.Context) {
		// 登录时的一大堆校验
		id := uuid.New()
		session, err := manager.InitSession(ctx, id.String())
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			return
		}
		// 然后根据自己的需要，设置值
		err = session.Set(ctx.Req.Context(), "mykey", "some value")
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			return
		}
		val, err := session.Get(ctx.Req.Context(), "mykey")
		ctx.RespData = []byte(val)
	})

	server.GET("/resource", func(ctx *web.Context) {
		session, err := manager.GetSession(ctx)
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			return
		}
		val, err := session.Get(ctx.Req.Context(), "mykey")
		ctx.RespData = []byte(val)
	})

	server.GET("/logout", func(ctx *web.Context) {
		_ = manager.RemoveSession(ctx)
	})
	server.Start(":8081")
}
