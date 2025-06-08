package main

import (
	"log"

	"github.com/chaikadn/url-shortener/internal/config"
	"github.com/chaikadn/url-shortener/internal/generator"
	"github.com/chaikadn/url-shortener/internal/handler"
	"github.com/chaikadn/url-shortener/internal/logger"
	"github.com/chaikadn/url-shortener/internal/server"
	"github.com/chaikadn/url-shortener/internal/service"
	"github.com/chaikadn/url-shortener/internal/storage/pg"

	"go.uber.org/zap"
)

func main() {
	cfg := config.New()
	if err := cfg.Load(); err != nil {
		log.Fatalf("failed to initialize config: %v", err)
	}

	zlog, err := logger.New(cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer zlog.Sync()

	stg, err := pg.New(zlog, cfg.DatabaseDSN)
	if err != nil {
		zlog.Fatal("failed to initialize storage", zap.Error(err))
	}
	defer stg.Close()

	shortenerService := service.NewShortenerService(stg, generator.NewCrypto())
	userService := service.NewUserService(stg)

	hnd := handler.New(zlog, cfg, shortenerService, userService)
	if err := server.New(zlog, cfg, hnd).Run(); err != nil {
		zlog.Fatal("server error", zap.Error(err))
	}
}
