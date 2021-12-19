package main

import (
	"imola/kernel"
	"net/http"
)

func main() {
	server := &http.Server{
		Handler: kernel.NewCore(),
		Addr:    "localhost:8888",
	}
	server.ListenAndServe()
}
