package server

import (
	"fmt"
	"net/http"

	"github.com/chaikadn/url-shortener/internal/app/handler"
	"github.com/chaikadn/url-shortener/internal/app/storage/memory"
)

// TODO: refactor

type Server struct {
	host string
	port string
}

func New(host, port string) *Server {
	return &Server{
		host: host,
		port: port,
	}
}

func (s *Server) Run() error {
	st := memory.NewStorage()
	hr := handler.New(st)

	// Add handlers here:
	http.HandleFunc("/", hr.ShortenURL)
	http.HandleFunc("/{id}", hr.GetURL)

	fmt.Println("service started")

	return http.ListenAndServe(s.host+s.port, nil)
}
