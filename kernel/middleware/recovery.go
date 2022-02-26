package middleware

import "imola/kernel"

// Recovery recovery机制，将协程中的函数异常进行捕获
func Recovery() kernel.ControllerHandler {
	return func(ctx *kernel.Context) error {
		// 核心在增加这个recovery()机制，捕获ctx.Next()出现的panic
		defer func() {
			if err := recover(); err != nil {
				ctx.Json(500, err)
			}
		}()
		ctx.Next()
		return nil
	}
}
