package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

const (
	DefaultNATSURL              = "nats://127.0.0.1:4222"
	DefaultNATSStreamName       = "CDC_EVENTS"
	DefaultNATSRetentionDays    = 7
	DefaultNATSMaxReconnects    = -1 // unlimited
	DefaultNATSReconnectWaitMs  = 2000
	DefaultNATSReconnectBufSize = 64
	DefaultNATSMaxAckPending    = 1000
	DefaultNATSAckWaitMs        = 30000
	DefaultNATSMaxDeliver       = 5

	DefaultGRPCPort             = 9090
	DefaultHTTPPort             = 9091
	DefaultPrometheusURL        = "http://127.0.0.1:9095"
	DefaultPrometheusQueryRange = "5m"

	DefaultLogMode  = "text"
	DefaultLogLevel = "info"
)

// Config holds server-level settings only. No sources, sinks, or flow config.
type Config struct {
	NATS       NATSConfig       `mapstructure:"nats" json:"nats"`
	Server     ServerConfig     `mapstructure:"server" json:"server"`
	Prometheus PrometheusConfig `mapstructure:"prometheus" json:"prometheus"`
	Log        LogConfig        `mapstructure:"log" json:"log"`
}

// NATSConfig holds the configuration for NATS JetStream connection.
type NATSConfig struct {
	URL                   string `mapstructure:"url" json:"url"`
	StreamName            string `mapstructure:"stream_name" json:"stream_name"`
	RetentionDays         int32  `mapstructure:"retention_days" json:"retention_days"`
	MaxReconnects         int    `mapstructure:"max_reconnects" json:"max_reconnects"`
	ReconnectWaitMs       int    `mapstructure:"reconnect_wait_ms" json:"reconnect_wait_ms"`
	ReconnectBufferSizeMb int    `mapstructure:"reconnect_buffer_size_mb" json:"reconnect_buffer_size_mb"`
	MaxAckPending         int    `mapstructure:"max_ack_pending" json:"max_ack_pending"`
	AckWaitMs             int    `mapstructure:"ack_wait_ms" json:"ack_wait_ms"`
	MaxDeliver            int    `mapstructure:"max_deliver" json:"max_deliver"`
}

// ServerConfig holds gRPC + REST gateway port configuration.
type ServerConfig struct {
	GRPCPort int `mapstructure:"grpc_port" json:"grpc_port"`
	HTTPPort int `mapstructure:"http_port" json:"http_port"`
}

type PrometheusConfig struct {
	URL         string `mapstructure:"url" json:"url"`
	QueryWindow string `mapstructure:"query_window" json:"query_window"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Mode       string `mapstructure:"mode" json:"mode"`                         // "json" | "text"
	Level      string `mapstructure:"level" json:"level"`                       // "debug" | "info" | "warn" | "error"
	FilePath   string `mapstructure:"file_path" json:"file_path,omitempty"`     // optional log file path
	MaxSizeMB  int    `mapstructure:"max_size_mb" json:"max_size_mb,omitempty"` // log rotation size
	MaxBackups int    `mapstructure:"max_backups" json:"max_backups,omitempty"` // rotated files to keep
}

// LoadConfig loads configuration from config.yaml with environment variable overrides.
// Environment variables: NATS_URL, GRPC_PORT, HTTP_PORT, LOG_MODE, LOG_LEVEL
// Defaults: nats://127.0.0.1:4222, 9090, 9091, "text", "info"
func LoadConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("config")
	v.AddConfigPath(".")

	// Enable environment variable reading
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind specific environment variables to config keys
	_ = v.BindEnv("nats.url", "NATS_URL")
	_ = v.BindEnv("server.grpc_port", "GRPC_PORT")
	_ = v.BindEnv("server.http_port", "HTTP_PORT")
	_ = v.BindEnv("prometheus.url", "PROMETHEUS_URL")
	_ = v.BindEnv("prometheus.query_window", "PROMETHEUS_QUERY_WINDOW")
	_ = v.BindEnv("log.mode", "LOG_MODE")
	_ = v.BindEnv("log.level", "LOG_LEVEL")

	// Read config file (non-fatal if missing — defaults will apply)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// For other errors (e.g., parse errors), return the error
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is OK — env vars and defaults will be used
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	applyDefaults(&cfg)

	return &cfg, nil
}

// applyDefaults sets default values for any fields not provided by file or env.
func applyDefaults(c *Config) {
	// NATS defaults
	if c.NATS.URL == "" {
		c.NATS.URL = DefaultNATSURL
	}
	if c.NATS.StreamName == "" {
		c.NATS.StreamName = DefaultNATSStreamName
	}
	if c.NATS.RetentionDays <= 0 {
		c.NATS.RetentionDays = DefaultNATSRetentionDays
	}
	if c.NATS.MaxReconnects == 0 {
		c.NATS.MaxReconnects = DefaultNATSMaxReconnects
	}
	if c.NATS.ReconnectWaitMs <= 0 {
		c.NATS.ReconnectWaitMs = DefaultNATSReconnectWaitMs
	}
	if c.NATS.ReconnectBufferSizeMb <= 0 {
		c.NATS.ReconnectBufferSizeMb = DefaultNATSReconnectBufSize
	}
	if c.NATS.MaxAckPending <= 0 {
		c.NATS.MaxAckPending = DefaultNATSMaxAckPending
	}
	if c.NATS.AckWaitMs <= 0 {
		c.NATS.AckWaitMs = DefaultNATSAckWaitMs
	}
	if c.NATS.MaxDeliver <= 0 {
		c.NATS.MaxDeliver = DefaultNATSMaxDeliver
	}

	// Server defaults
	if c.Server.GRPCPort <= 0 {
		c.Server.GRPCPort = DefaultGRPCPort
	}
	if c.Server.HTTPPort <= 0 {
		c.Server.HTTPPort = DefaultHTTPPort
	}
	if c.Prometheus.URL == "" {
		c.Prometheus.URL = DefaultPrometheusURL
	}
	if c.Prometheus.QueryWindow == "" {
		c.Prometheus.QueryWindow = DefaultPrometheusQueryRange
	}

	// Log defaults
	if c.Log.Mode == "" {
		c.Log.Mode = DefaultLogMode
	}
	if c.Log.Level == "" {
		c.Log.Level = DefaultLogLevel
	}
}
