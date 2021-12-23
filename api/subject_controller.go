package api

import "imola/kernel"

func SubjectAddController(ctx *kernel.Context) error {
	ctx.Json(200, "ok, SubjectAddController")
	return nil
}
