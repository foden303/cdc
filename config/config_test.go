package config

import (
	"os"
	"testing"
)

func TestApplyDefaults(t *testing.T) {
	cfg := &Config{}
	applyDefaults(cfg)

	if cfg.NATS.URL != "nats://127.0.0.1:4222" {
		t.Errorf("expected NATS URL default nats://127.0.0.1:4222, got %s", cfg.NATS.URL)
	}
	if cfg.NATS.StreamName != "CDC_EVENTS" {
		t.Errorf("expected NATS StreamName default CDC_EVENTS, got %s", cfg.NATS.StreamName)
	}
	if cfg.NATS.RetentionDays != 7 {
		t.Errorf("expected NATS RetentionDays default 7, got %d", cfg.NATS.RetentionDays)
	}
	if cfg.NATS.MaxReconnects != -1 {
		t.Errorf("expected NATS MaxReconnects default -1, got %d", cfg.NATS.MaxReconnects)
	}
	if cfg.NATS.ReconnectWaitMs != 2000 {
		t.Errorf("expected NATS ReconnectWaitMs default 2000, got %d", cfg.NATS.ReconnectWaitMs)
	}
	if cfg.NATS.ReconnectBufferSizeMb != 64 {
		t.Errorf("expected NATS ReconnectBufferSizeMb default 64, got %d", cfg.NATS.ReconnectBufferSizeMb)
	}
	if cfg.Server.GRPCPort != 9090 {
		t.Errorf("expected Server GRPCPort default 9090, got %d", cfg.Server.GRPCPort)
	}
	if cfg.Server.HTTPPort != 9091 {
		t.Errorf("expected Server HTTPPort default 9091, got %d", cfg.Server.HTTPPort)
	}
	if cfg.Log.Mode != "text" {
		t.Errorf("expected Log Mode default text, got %s", cfg.Log.Mode)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("expected Log Level default info, got %s", cfg.Log.Level)
	}
}

func TestApplyDefaults_PreservesExistingValues(t *testing.T) {
	cfg := &Config{
		NATS: NATSConfig{
			URL:        "nats://custom:4222",
			StreamName: "CUSTOM_STREAM",
		},
		Server: ServerConfig{
			GRPCPort: 8080,
			HTTPPort: 8081,
		},
		Log: LogConfig{
			Mode:  "json",
			Level: "debug",
		},
	}
	applyDefaults(cfg)

	if cfg.NATS.URL != "nats://custom:4222" {
		t.Errorf("expected NATS URL to be preserved, got %s", cfg.NATS.URL)
	}
	if cfg.NATS.StreamName != "CUSTOM_STREAM" {
		t.Errorf("expected NATS StreamName to be preserved, got %s", cfg.NATS.StreamName)
	}
	if cfg.Server.GRPCPort != 8080 {
		t.Errorf("expected Server GRPCPort to be preserved, got %d", cfg.Server.GRPCPort)
	}
	if cfg.Server.HTTPPort != 8081 {
		t.Errorf("expected Server HTTPPort to be preserved, got %d", cfg.Server.HTTPPort)
	}
	if cfg.Log.Mode != "json" {
		t.Errorf("expected Log Mode to be preserved, got %s", cfg.Log.Mode)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("expected Log Level to be preserved, got %s", cfg.Log.Level)
	}
}

func TestLoadConfig_EnvVarOverrides(t *testing.T) {
	// Set environment variables
	os.Setenv("NATS_URL", "nats://envhost:4222")
	os.Setenv("GRPC_PORT", "7070")
	os.Setenv("HTTP_PORT", "7071")
	os.Setenv("LOG_MODE", "json")
	os.Setenv("LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("NATS_URL")
		os.Unsetenv("GRPC_PORT")
		os.Unsetenv("HTTP_PORT")
		os.Unsetenv("LOG_MODE")
		os.Unsetenv("LOG_LEVEL")
	}()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.NATS.URL != "nats://envhost:4222" {
		t.Errorf("expected NATS URL from env nats://envhost:4222, got %s", cfg.NATS.URL)
	}
	if cfg.Server.GRPCPort != 7070 {
		t.Errorf("expected GRPC port from env 7070, got %d", cfg.Server.GRPCPort)
	}
	if cfg.Server.HTTPPort != 7071 {
		t.Errorf("expected HTTP port from env 7071, got %d", cfg.Server.HTTPPort)
	}
	if cfg.Log.Mode != "json" {
		t.Errorf("expected Log Mode from env json, got %s", cfg.Log.Mode)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("expected Log Level from env debug, got %s", cfg.Log.Level)
	}
}

func TestLoadConfig_DefaultsFromYAML(t *testing.T) {
	// Clear any env vars that might interfere
	os.Unsetenv("NATS_URL")
	os.Unsetenv("GRPC_PORT")
	os.Unsetenv("HTTP_PORT")
	os.Unsetenv("LOG_MODE")
	os.Unsetenv("LOG_LEVEL")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// These should come from config.yaml
	if cfg.NATS.URL != "nats://127.0.0.1:4222" {
		t.Errorf("expected NATS URL nats://127.0.0.1:4222, got %s", cfg.NATS.URL)
	}
	if cfg.NATS.StreamName != "CDC_EVENTS" {
		t.Errorf("expected NATS StreamName CDC_EVENTS, got %s", cfg.NATS.StreamName)
	}
	if cfg.Server.GRPCPort != 9090 {
		t.Errorf("expected GRPC port 9090, got %d", cfg.Server.GRPCPort)
	}
	if cfg.Server.HTTPPort != 9091 {
		t.Errorf("expected HTTP port 9091, got %d", cfg.Server.HTTPPort)
	}
	if cfg.Log.Mode != "text" {
		t.Errorf("expected Log Mode text, got %s", cfg.Log.Mode)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("expected Log Level info, got %s", cfg.Log.Level)
	}
}
