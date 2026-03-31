// Package config handles loading and validating service configuration.
package config

import (
	"log/slog"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Database    *DBConfig
	Service     *ServiceConfig
	HealthCheck *HealthCheckConfig
	NATS        *NATSConfig
	Cleanup     *CleanupConfig
}

type CleanupConfig struct {
	Interval   time.Duration `envconfig:"CLEANUP_INTERVAL" default:"1m"`
	MaxRetries int           `envconfig:"CLEANUP_MAX_RETRIES" default:"10"`
	Timeout    time.Duration `envconfig:"CLEANUP_TIMEOUT" default:"10s"`
}

type HealthCheckConfig struct {
	Interval               time.Duration `envconfig:"HEALTH_CHECK_INTERVAL" default:"10s"`
	Timeout                time.Duration `envconfig:"HEALTH_CHECK_TIMEOUT" default:"5s"`
	MaxConsecutiveFailures int           `envconfig:"HEALTH_CHECK_MAX_CONSECUTIVE_FAILURES" default:"3"`
	BaseBackoffInterval    time.Duration `envconfig:"HEALTH_CHECK_BASE_BACKOFF_INTERVAL" default:"10s"`
	MaxBackoffInterval     time.Duration `envconfig:"HEALTH_CHECK_MAX_BACKOFF_INTERVAL" default:"5m"`
}

type DBConfig struct {
	Type     string `envconfig:"DB_TYPE" default:"pgsql"`
	Hostname string `envconfig:"DB_HOST" default:"localhost"`
	Port     string `envconfig:"DB_PORT" default:"5432"`
	Name     string `envconfig:"DB_NAME" default:"service-provider"`
	User     string `envconfig:"DB_USER"`
	Password string `envconfig:"DB_PASS"`
}

type ServiceConfig struct {
	Address  string `envconfig:"SVC_ADDRESS" default:":8080"`
	LogLevel string `envconfig:"SVC_LOG_LEVEL" default:"info"`
}

// NATSConfig holds NATS messaging configuration
type NATSConfig struct {
	URL          string `envconfig:"NATS_URL" default:"nats://localhost:4222"`
	Subject      string `envconfig:"NATS_STATUS_SUBJECT" default:"dcm.*"`
	StreamName   string `envconfig:"NATS_STREAM_NAME" default:"dcm-status"`
	ConsumerName string `envconfig:"NATS_CONSUMER_NAME" default:"service-provider-manager"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}
	if cfg.Database.Type != "pgsql" && cfg.Database.Type != "sqlite" {
		slog.Warn("Invalid DB_TYPE, defaulting to sqlite", "db_type", cfg.Database.Type)
		cfg.Database.Type = "sqlite"
	}
	return cfg, nil
}
