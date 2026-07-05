package config

import (
	"strings"
	"time"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Callback struct {
		Addr    string        `env:"CALLBACK_ADDR"`
		Timeout time.Duration `env:"CALLBACK_TIMEOUT" envDefault:"5s"`
	}

	GRPC struct {
		Port string `env:"GRPC_PORT" envDefault:"50053"`
	}
}

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	return &cfg, err
}

func (c *Config) CallbackURL() string {
	addr := strings.TrimSpace(c.Callback.Addr)
	if addr == "" {
		return ""
	}
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return addr
	}
	return "http://" + addr
}
