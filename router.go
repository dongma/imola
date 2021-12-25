package main

import (
	"imola/api"
	"imola/kernel"
)

func registerRouter(core *kernel.Core) {
	// 需求1+2: HTTP方法+静态路由匹配
	core.Get("/user/login", api.UserLoginController)
	// 需求3: 批量通用前缀/subject
	subjectApi := core.Group("/subject")
	{
		// 需求4: 动态路由
		subjectApi.Delete("/:id", api.SubjectDelController)
		subjectApi.Put("/:id", api.SubjectUpdateController)
		subjectApi.Get("/:id", api.SubjectAddController)
		subjectApi.Get("/list/all", api.SubjectListController)

		subjectInnerApi := subjectApi.Group("/info")
		{
			subjectInnerApi.Get("/name", api.SubjectNameController)
		}
	}
}
