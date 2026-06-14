package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}

func TestLoad_FullConfig(t *testing.T) {
	content := `
server:
  port: 9090
  host: "127.0.0.1"
database:
  url: "postgres://user:pass@localhost:5432/bsl"
  max_connections: 10
bilibili:
  cookie: "SESSDATA=xxx"
onebot:
  ws_path: "/custom/onebot"
collector:
  heartbeat_interval: 15
  reconnect_backoff_max: 120
  stats_flush_interval: 30
`
	cfg, err := Load(writeTemp(t, content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d, want 9090", cfg.Server.Port)
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host = %s, want 127.0.0.1", cfg.Server.Host)
	}
	if cfg.Database.URL != "postgres://user:pass@localhost:5432/bsl" {
		t.Errorf("Database.URL = %s", cfg.Database.URL)
	}
	if cfg.Database.MaxConnections != 10 {
		t.Errorf("Database.MaxConnections = %d, want 10", cfg.Database.MaxConnections)
	}
	if cfg.Bilibili.Cookie != "SESSDATA=xxx" {
		t.Errorf("Bilibili.Cookie = %s", cfg.Bilibili.Cookie)
	}
	if cfg.OneBot.WSPath != "/custom/onebot" {
		t.Errorf("OneBot.WSPath = %s", cfg.OneBot.WSPath)
	}
	if cfg.Collector.HeartbeatInterval != 15 {
		t.Errorf("HeartbeatInterval = %d, want 15", cfg.Collector.HeartbeatInterval)
	}
	if cfg.Collector.ReconnectBackoffMax != 120 {
		t.Errorf("ReconnectBackoffMax = %d, want 120", cfg.Collector.ReconnectBackoffMax)
	}
	if cfg.Collector.StatsFlushInterval != 30 {
		t.Errorf("StatsFlushInterval = %d, want 30", cfg.Collector.StatsFlushInterval)
	}
}

func TestLoad_Defaults(t *testing.T) {
	content := `
database:
  url: "postgres://localhost/bsl"
`
	cfg, err := Load(writeTemp(t, content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("default Server.Port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("default Server.Host = %s, want 0.0.0.0", cfg.Server.Host)
	}
	if cfg.Database.MaxConnections != 20 {
		t.Errorf("default MaxConnections = %d, want 20", cfg.Database.MaxConnections)
	}
	if cfg.Collector.HeartbeatInterval != 30 {
		t.Errorf("default HeartbeatInterval = %d, want 30", cfg.Collector.HeartbeatInterval)
	}
	if cfg.Collector.ReconnectBackoffMax != 60 {
		t.Errorf("default ReconnectBackoffMax = %d, want 60", cfg.Collector.ReconnectBackoffMax)
	}
	if cfg.Collector.StatsFlushInterval != 60 {
		t.Errorf("default StatsFlushInterval = %d, want 60", cfg.Collector.StatsFlushInterval)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	content := `server: [bad: yaml`
	_, err := Load(writeTemp(t, content))
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoad_PartialConfig(t *testing.T) {
	content := `
server:
  port: 3000
database:
  url: "postgres://localhost/bsl"
`
	cfg, err := Load(writeTemp(t, content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 3000 {
		t.Errorf("Server.Port = %d, want 3000", cfg.Server.Port)
	}
	// Host should use default
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %s, want 0.0.0.0 (default)", cfg.Server.Host)
	}
}
