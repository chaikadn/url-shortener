package config

import "flag"

type Config struct {
	Host    string
	BaseURL string
}

func New() *Config {
	return &Config{}
}

func (c *Config) ParseFlags() {
	flag.StringVar(&c.Host, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&c.BaseURL, "b", "http://localhost:8080", "base short URL addres")

	flag.Parse()
}
