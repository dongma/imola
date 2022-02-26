package middleware

import (
	"fmt"
	"imola/kernel"
)

func Test1() kernel.ControllerHandler {
	// 使用函数进行回调，调用Next往下调用，自增context, index
	return func(ctx *kernel.Context) error {
		fmt.Println("middleware are test1")
		ctx.Next()
		fmt.Println("middleware post test1")
		return nil
	}
}

func Test2() kernel.ControllerHandler {
	return func(ctx *kernel.Context) error {
		fmt.Println("middleware are test2")
		ctx.Next()
		fmt.Println("middleware post test2")
		return nil
	}
}
