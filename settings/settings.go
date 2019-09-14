package settings

import (
	"context"
	"time"
)

type ctxKey int

const (
	settingsKey ctxKey = iota
	loggerKey
	requestInfoKey
)

// Get returns settings from the context.
func Get(ctx context.Context) *Settings {
	return ctx.Value(settingsKey).(*Settings)
}

// Set returns a new context with set (or re-set) settings.
func Set(ctx context.Context, s *Settings) context.Context {
	return context.WithValue(ctx, settingsKey, s)
}

type Settings struct {
	Postgres PostgresConfig `json:"postgres"`

	API ApiConfig `json:"api"`

	Debug      DebugConfig      `json:"debug"`
	Sentry     SentryConfig     `json:"sentry"`
	Prometheus PrometheusConfig `json:"prometheus_config"`
}

type PrometheusConfig struct {
	MetricsListenAddress string `json:"metrics_listen_address"`
	RegSerivceName       string `json:"reg_service_name"`
}

type ApiConfig struct {
	GRPCListenAddress    string `json:"grpc_listen_address"`
	GRPCWebListenAddress string `json:"grpc_web_listen_address"`
	RESTListenAddress    string `json:"rest_listen_address"`
}

type DebugConfig struct {
	DebugEmailPattern       string `json:"debug_email_pattern"`
	EnableReflection        bool   `json:"enable_reflection"`
	EnableDevelopmentLogger bool   `json:"enable_development_logger"`
	EnableDebugLevelLogger  bool   `json:"enable_debug_level_logger"`
}

type PostgresConfig struct {
	URL             string        `json:"url"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
}

type SentryConfig struct {
	DSN string `json:"dsn"`
}
