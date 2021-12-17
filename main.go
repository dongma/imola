package main

import (
	"imola/core"
	"net/http"
)

func main() {
	core := core.NewCore()
	server := &http.Server{
		Handler: core,
		Addr:    ":8888",
	}
	server.ListenAndServe()
}
