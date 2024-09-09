package main

import "github.com/chaikadn/url-shortener/internal/app/server"

func main() {

	srv := server.New("", ":8080")

	if err := srv.Run(); err != nil {
		panic(err)
	}
}
