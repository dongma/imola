package opentelemetry

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"imola"
)

const instrumentationName = "/imola/middlewares/opentelemetry"

type MiddlewareBuilder struct {
	Tracer trace.Tracer
}

func (m MiddlewareBuilder) Build() imola.Middleware {
	if m.Tracer == nil {
		m.Tracer = otel.GetTracerProvider().Tracer(instrumentationName)
	}
	return func(next imola.HandleFunc) imola.HandleFunc {
		return func(ctx *imola.Context) {
			reqCtx := ctx.Req.Context()
			// 尝试和客户端的trace结合在一起
			reqCtx = otel.GetTextMapPropagator().Extract(reqCtx, propagation.HeaderCarrier(ctx.Req.Header))

			_, span := m.Tracer.Start(reqCtx, "unknown")
			defer span.End()
			span.SetAttributes(attribute.String("http.method", ctx.Req.Method))
			span.SetAttributes(attribute.String("http.url", ctx.Req.URL.String()))
			span.SetAttributes(attribute.String("http.method", ctx.Req.URL.Scheme))
			span.SetAttributes(attribute.String("http.host", ctx.Req.Host))

			// 直接调用下一步，这个只有执行完next才可能有值
			next(ctx)
			span.SetName(ctx.MatchedRoute)
		}
	}
}
