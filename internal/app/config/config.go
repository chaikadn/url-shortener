package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Host            string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	LogLevel        string `env:"LOG_LEVEL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

func New() *Config {
	return &Config{
		Host:            "localhost:8080",
		BaseURL:         "http://localhost:8080",
		LogLevel:        "info",
		FileStoragePath: "",

		// "postgresql://user:password@localhost:5432/mydb?sslmode=disable" - для pgx
		// "host=localhost port=5432 user=user password=password dbname=mydb sslmode=disable" - для lib/pq
		DatabaseDSN: "",
	}
}

func (c *Config) Load() error {
	c.parseFlags()
	if err := c.parseEnv(); err != nil {
		return err
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
