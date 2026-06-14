package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	Bilibili  BilibiliConfig  `yaml:"bilibili"`
	OneBot    OneBotConfig    `yaml:"onebot"`
	Collector CollectorConfig `yaml:"collector"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type DatabaseConfig struct {
	URL            string `yaml:"url"`
	MaxConnections int    `yaml:"max_connections"`
}

type BilibiliConfig struct {
	Cookie string `yaml:"cookie"`
}

type OneBotConfig struct {
	WSPath string `yaml:"ws_path"`
}

type CollectorConfig struct {
	HeartbeatInterval   int `yaml:"heartbeat_interval"`
	ReconnectBackoffMax int `yaml:"reconnect_backoff_max"`
	StatsFlushInterval  int `yaml:"stats_flush_interval"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	setDefaults(cfg)
	return cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Database.MaxConnections == 0 {
		cfg.Database.MaxConnections = 20
	}
	if cfg.Collector.HeartbeatInterval == 0 {
		cfg.Collector.HeartbeatInterval = 30
	}
	if cfg.Collector.ReconnectBackoffMax == 0 {
		cfg.Collector.ReconnectBackoffMax = 60
	}
	if cfg.Collector.StatsFlushInterval == 0 {
		cfg.Collector.StatsFlushInterval = 60
	}
}
