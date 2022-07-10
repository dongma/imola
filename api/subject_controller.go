package api

import "imola/kernel"

func SubjectAddController(ctx *kernel.Context) error {
	ctx.Json(200, "ok, SubjectAddController")
	return nil
}

func SubjectListController(c *kernel.Context) error {
	c.Json(200, "ok, SubjectListController")
	return nil
}

func SubjectDelController(c *kernel.Context) error {
	c.Json(200, "ok, SubjectDelController")
	return nil
}

func SubjectUpdateController(c *kernel.Context) error {
	c.Json(200, "ok, SubjectUpdateController")
	return nil
}

func SubjectGetController(c *kernel.Context) error {
	c.Json(200, "ok, SubjectGetController")
	return nil
}

func SubjectNameController(c *kernel.Context) error {
	c.Json(200, "ok, SubjectNameController")
	return nil
}
