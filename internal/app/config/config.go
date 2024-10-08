package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Host    string `env:"SERVER_ADDRESS"`
	BaseURL string `env:"BASE_URL"`
}

func New() *Config {
	return &Config{
		Host:    "localhost:8080",
		BaseURL: "http://localhost:8080",
	}
}

func (c *Config) ParseFlags() {
	flag.StringVar(&c.Host, "a", c.Host, "address and port to run server")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "base short URL addres")
	flag.Parse()
}

func (c *Config) ParseEnv() {
	env.Parse(c)
}
