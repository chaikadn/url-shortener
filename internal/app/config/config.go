package config

import (
	"errors"
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	Host            string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	LogLevel        string `env:"LOG_LEVEL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	JWTSecret       string `env:"JWT_SECRET"`
}

func New() *Config {
	return &Config{
		Host:      "localhost:8080",
		BaseURL:   "http://localhost:8080",
		LogLevel:  "info",
		JWTSecret: "temporary",
	}
}

func (c *Config) Load() error {
	// только для разработки
	if err := godotenv.Load(); err != nil {
		log.Printf("\t\tfailed to load .env file: %v", err)
	} else {
		log.Printf("\t\t.env file loaded successfully")
	}

	c.parseFlags()
	if err := c.parseEnv(); err != nil {
		return err
	}

	if c.JWTSecret == "" {
		return errors.New("JWT_SECRET must be set")
	}

	return nil
}

func (c *Config) parseFlags() {
	flag.StringVar(&c.Host, "a", c.Host, "address and port to run server")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "base short URL address")
	flag.StringVar(&c.LogLevel, "l", c.LogLevel, "log level")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "file storage path")
	flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "database DSN in URI format")
	flag.Parse()
}

func (c *Config) parseEnv() error {
	if err := env.Parse(c); err != nil {
		return err
	}
	return nil
}
