package kernel

import (
	"context"
	"fmt"
	"log"
	"time"
)

func TimeoutHandler(fun ControllerHandler, duration time.Duration) ControllerHandler {
	// 使用函数回调
	return func(ctx *Context) error {
		finish := make(chan struct{}, 1)
		panicChan := make(chan interface{}, 1)

		// 执行业务逻辑前的预操作，初始化超时context
		durationCtx, cancel := context.WithTimeout(ctx.BaseContext(), duration)
		defer cancel()

		ctx.request.WithContext(ctx)

		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()
			// 执行具体业务逻辑
			fun(ctx)
			finish <- struct{}{}
		}()

		// 执行业务逻辑后操作
		select {
		case p := <-panicChan:
			log.Println(p)
			ctx.responseWriter.WriteHeader(500)
		case <-finish:
			fmt.Println("finish")
		case <-durationCtx.Done():
			ctx.SetHasTimeout()
			ctx.responseWriter.Write([]byte("time out"))
		}
		return nil
	}
}

// Timeout 新版本的Timeout，超时控制参数中ControllerHandler已经去掉
func Timeout(duration time.Duration) ControllerHandler {
	return func(ctx *Context) error {
		// 使用callback function返回数据
		return nil
	}
}
