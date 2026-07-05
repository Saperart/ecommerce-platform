package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "postgres")
	t.Setenv("POSTGRES_PORT", "5433")
	t.Setenv("POSTGRES_DB", "shop")
	t.Setenv("POSTGRES_USER", "user@example.com")
	t.Setenv("POSTGRES_PASSWORD", "pass word")
	t.Setenv("NOTIFICATIONS_GRPC_ADDR", "notifications:50053")
	t.Setenv("OUTBOX_WORKERS", "2")
	t.Setenv("OUTBOX_BATCH_SIZE", "20")
	t.Setenv("OUTBOX_FETCH_PERIOD", "300ms")
	t.Setenv("OUTBOX_IN_PROGRESS_TTL", "11s")
	t.Setenv("GRPC_PORT", "50099")
	t.Setenv("GRPC_GATEWAY_PORT", "8099")

	cfg, err := New()

	require.NoError(t, err)
	require.Equal(t, "postgres", cfg.PG.Host)
	require.Equal(t, "5433", cfg.PG.Port)
	require.Equal(t, "shop", cfg.PG.DB)
	require.Equal(t, "user@example.com", cfg.PG.User)
	require.Equal(t, "pass word", cfg.PG.Password)
	require.Equal(t, "notifications:50053", cfg.Clients.NotificationsGrpcAddr)
	require.Equal(t, 2, cfg.Outbox.Workers)
	require.Equal(t, 20, cfg.Outbox.BatchSize)
	require.Equal(t, 300*time.Millisecond, cfg.Outbox.FetchPeriod)
	require.Equal(t, 11*time.Second, cfg.Outbox.InProgressTTL)
	require.Equal(t, "50099", cfg.GRPC.Port)
	require.Equal(t, "8099", cfg.GRPC.GatewayPort)
}

func TestConfigPostgresDSN(t *testing.T) {
	t.Parallel()

	cfg := &Config{}
	cfg.PG.Host = "localhost"
	cfg.PG.Port = "5432"
	cfg.PG.DB = "ecommerce"
	cfg.PG.User = "user@example.com"
	cfg.PG.Password = "pass word"

	dsn := cfg.PostgresDSN()

	require.Equal(t, "postgres://user%40example.com:pass+word@localhost:5432/ecommerce?sslmode=disable", dsn)
}
