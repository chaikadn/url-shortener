package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	Host        string `env:"SERVER_HOST"`
	Port        string `env:"SERVER_PORT"`
	LogLevel    string `env:"LOG_LEVEL"`
	DatabaseDSN string `env:"DATABASE_DSN"`
	JWTsecret   string `env:"JWT_SECRET"`
}

func New() *Config {
	return &Config{
		Host:        "localhost",
		Port:        "8080",
		LogLevel:    "info",
		DatabaseDSN: "",
	}
}

func (c *Config) Load() error {
	// for dev only
	if err := c.loadDotEnv(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("load .env: %w", err)
	}

	if err := c.parseEnv(); err != nil {
		return fmt.Errorf("parse env: %w", err)
	}

	c.parseFlags()

	if c.JWTsecret == "" {
		return fmt.Errorf("load config: JWT_SECRET is not set")
	}

	log.Printf("\t\t\tConfig loaded successfully")
	return nil
}

func (c *Config) parseFlags() {
	flag.StringVar(&c.Host, "h", c.Host, "server host")
	flag.StringVar(&c.Port, "p", c.Port, "server port")
	flag.StringVar(&c.LogLevel, "l", c.LogLevel, "log level")
	flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "database DSN in URI format")

	flag.Parse()
}

func (c *Config) parseEnv() error {
	if err := env.Parse(c); err != nil {
		return fmt.Errorf("failed to parse env: %w", err)
	}
	return nil
}

func (c *Config) loadDotEnv() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}
	return nil
}
