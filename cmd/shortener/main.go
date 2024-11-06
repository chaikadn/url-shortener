package main

import (
	"github.com/chaikadn/url-shortener/internal/app/config"
	"github.com/chaikadn/url-shortener/internal/app/handler"
	"github.com/chaikadn/url-shortener/internal/app/logger"
	"github.com/chaikadn/url-shortener/internal/app/server"
	"github.com/chaikadn/url-shortener/internal/app/storage/memory"
	"go.uber.org/zap"
)

func main() {
	config := config.New()
	config.ParseFlags()
	config.ParseEnv()

	if err := logger.Initialize(config.LogLevel); err != nil {
		panic("cannot initialize loger: " + err.Error())
	}
	defer logger.Log.Sync()

	storage := memory.NewStorage()
	handler := handler.New(storage, config)
	server := server.New(handler, config)

	logger.Log.Info("Starting server", zap.String("host", config.Host))
	if err := server.ListenAndServe(); err != nil {
		logger.Log.Fatal("cannot start server: " + err.Error())
	}
}
