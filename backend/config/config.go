package config

import (
	"encoding/json"
	"errors"
	"net"
	"os"
	"strconv"
)

// Config holds the listener address and SQLite path.
type Config struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	DbPath  string `json:"db"`
}

// DefaultConfig returns local development settings.
func DefaultConfig() Config {
	return Config{
		Address: "localhost",
		Port:    8080,
		DbPath:  "data/local.sqlite",
	}
}

// New loads a JSON config over the defaults and validates its settings.
func New(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Address == "" || cfg.Port < 1 || cfg.Port > 65535 || cfg.DbPath == "" {
		return nil, errors.New("config requires address, port (1-65535), and db")
	}

	return &cfg, nil
}

// ListenAddress joins the configured host and port.
func (c Config) ListenAddress() string {
	return net.JoinHostPort(c.Address, strconv.Itoa(c.Port))
}
