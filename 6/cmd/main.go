package main

import (
	"homework6/internal/httpapi"
	"homework6/internal/model"
	"log"
	"net/http"
	"time"
)

func main() {
	server := &http.Server{
		Addr: "127.0.0.1:8080", Handler: httpapi.NewRouter(model.NewStore()),
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       10 * time.Second, WriteTimeout: 10 * time.Second,
		IdleTimeout: 60 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
