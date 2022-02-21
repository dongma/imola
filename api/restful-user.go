package api

import "imola/kernel"

func UserLoginController(ctx *kernel.Context) error {
	ctx.Json(200, "ok, UserLoginController")
	return nil
}
