package main

import (
	"imola/api"
	"imola/kernel"
)

func registerRouter(core *kernel.Core) {
	// 需求1+2: HTTP方法+静态路由匹配
	core.Get("/user/login", api.UserLoginController)
	subjectApi = core.Group("/subject")
	{
		subjectApi.Get("/add", api.SubjectAddController)
	}
}
