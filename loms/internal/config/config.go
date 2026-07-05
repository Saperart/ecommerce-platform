package config

import (
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/caarlos0/env/v10"
)

type (
	Config struct {
		PG struct {
			Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
			Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
			DB       string `env:"POSTGRES_DB" envDefault:"ecommerce"`
			User     string `env:"POSTGRES_USER" envDefault:"ecommerce_user"`
			Password string `env:"POSTGRES_PASSWORD" envDefault:"12345"`
		}

		Clients struct {
			NotificationsGrpcAddr string `env:"NOTIFICATIONS_GRPC_ADDR" envDefault:"localhost:50053"`
		}

		Outbox struct {
			Workers       int           `env:"OUTBOX_WORKERS" envDefault:"1"`
			BatchSize     int           `env:"OUTBOX_BATCH_SIZE" envDefault:"10"`
			FetchPeriod   time.Duration `env:"OUTBOX_FETCH_PERIOD" envDefault:"200ms"`
			InProgressTTL time.Duration `env:"OUTBOX_IN_PROGRESS_TTL" envDefault:"10s"`
		}

		GRPC struct {
			Port        string `env:"GRPC_PORT" envDefault:"50052"`
			GatewayPort string `env:"GRPC_GATEWAY_PORT" envDefault:"8081"`
		}
	}
)

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	return &cfg, err
}

func (c *Config) PostgresDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		url.QueryEscape(c.PG.User),
		url.QueryEscape(c.PG.Password),
		net.JoinHostPort(c.PG.Host, c.PG.Port),
		c.PG.DB,
	)
}
