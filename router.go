package main

import (
	"imola/kernel"
)

func registerRouter(core *kernel.Core) {
	// core.Get("foo", kernel.TimeoutHandler(FooControllerHandler, time.Second*1))
	core.Get("foo", FooControllerHandler)
}