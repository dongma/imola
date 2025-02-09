package imola

import (
	"context"
)

type TemplateEngine interface {
	// Render 渲染页面，tplName模版名称、data 渲染页面用的数据
	Render(ctx context.Context, tplName string, data any) ([]byte, error)
}
