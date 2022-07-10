package main

import (
	"context"
	"fmt"
	"imola/kernel"
	"log"
	"time"
)

func FooControllerHandler(c *kernel.Context) error {
	finish := make(chan struct{}, 1)
	panicChan := make(chan interface{}, 1)

	// 在业务处理前，创建具有定时功能的context
	durationCtx, cancel := context.WithTimeout(c.BaseContext(), time.Duration(1*time.Second))
	defer cancel()

	// mu := sync.Mutex{}
	go func() {
		defer func() {
			if p := recover(); p != nil {
				panicChan <- p
			}
		}()
		// Do real action，执行真正的业务操作
		time.Sleep(10 * time.Second)
		c.Json(200, "ok")
		finish <- struct{}{}
	}()

	// 使用select关键字，监听3种事件：异常事件、结束事件和超时事件。
	select {
	case p := <-panicChan:
		c.WriterMux().Lock()
		defer c.WriterMux().Unlock()
		log.Println(p)
		c.Json(500, "panic")
	case <-finish:
		fmt.Println("finish")
	case <-durationCtx.Done():
		c.WriterMux().Lock()
		defer c.WriterMux().Unlock()
		c.Json(500, "time out")
		c.SetHasTimeout()
	}
	return nil
}
