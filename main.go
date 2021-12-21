package main

import (
	"imola/kernel"
	"net/http"
)

func main() {
	core := kernel.NewCore()
	registerRouter(core)
	server := &http.Server{
		Handler: core,
		Addr:    "localhost:8888",
	}
	server.ListenAndServe()
}
