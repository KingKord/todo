// Package config загружает конфигурацию приложения из YAML.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config описывает конфигурацию сервиса.
type Config struct {
	GRPCAddr    string `yaml:"grpc_addr"`
	PostgresDSN string `yaml:"postgres_dsn"`
}

// Load читает YAML-конфигурацию по указанному пути.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}
