package main

import (
	"fmt"

	"github.com/chaikadn/url-shortener/internal/app/handler"
	"github.com/chaikadn/url-shortener/internal/app/server"
	"github.com/chaikadn/url-shortener/internal/app/storage/memory"
)

func main() {

	storage := memory.NewStorage()
	handler := handler.New(storage)
	server := server.New(*handler)

	fmt.Println("service started")
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
