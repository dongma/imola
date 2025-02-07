package recover

import "imola"

type MiddlewareBuilder struct {
	StatusCode int
	Data       []byte
	Log        func(ctx *imola.Context)
}

func (m MiddlewareBuilder) Build() imola.Middleware {
	return func(next imola.HandleFunc) imola.HandleFunc {
		return func(ctx *imola.Context) {
			defer func() {
				if err := recover(); err != nil {
					ctx.RespData = m.Data
					ctx.RespStatusCode = m.StatusCode
					m.Log(ctx)
				}
			}()
			next(ctx)
		}
	}
}
