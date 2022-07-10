package main

import (
	"imola/api"
	"imola/kernel"
	"imola/kernel/middleware"
)

func registerRouter(core *kernel.Core) {
	// 需求1+2: HTTP方法+静态路由匹配，在核心的业务逻辑之外，封装一层TimeoutHandler
	core.Get("/user/login", middleware.Test2(), api.UserLoginController)
	// 需求3: 批量通用前缀/subject
	subjectApi := core.Group("/subject")
	{
		// 需求4: 动态路由
		subjectApi.Delete("/:id", middleware.Test2(), api.SubjectDelController)
		subjectApi.Put("/:id", api.SubjectUpdateController)
		subjectApi.Get("/:id", api.SubjectGetController)
		subjectApi.Get("/list/all", api.SubjectListController)

		subjectInnerApi := subjectApi.Group("/info")
		{
			subjectInnerApi.Get("/name", api.SubjectNameController)
		}
	}
}
