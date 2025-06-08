package main

import (
	"log"

	"github.com/chaikadn/url-shortener/internal/logger"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const (
	addr = "http://localhost:8080/test"
)

func main() {
	zlog, err := logger.New("debug")

	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	client := resty.New()
	resp, err := client.R().
		SetHeader("Accept-Encoding", "gzip").
		Get(addr)
	if err != nil {
		zlog.Fatal("failed to send request", zap.Error(err))
	}

	zlog.Info(
		"Http response",
		zap.Int("size", len(resp.Body())),
		zap.String("Content-Encoding", resp.Header().Get("Content-Encoding")), // Проверьте заголовок
	)
}
