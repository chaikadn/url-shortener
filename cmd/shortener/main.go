package main

import (
	"fmt"

	"github.com/chaikadn/url-shortener/internal/app/config"
	"github.com/chaikadn/url-shortener/internal/app/handler"
	"github.com/chaikadn/url-shortener/internal/app/server"
	"github.com/chaikadn/url-shortener/internal/app/storage/memory"
)

func main() {
	config := config.New()
	config.ParseFlags()
	config.ParseEnv()

	storage := memory.NewStorage()
	handler := handler.New(storage, config)
	server := server.New(handler, config)

	fmt.Println("service started at", config.Host)
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
