package config

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Host               string `env:"HOST" envDefault:"http://localhost:8080"`
	GeniusClientID     string `env:"GENIUS_API_ID"`
	GeniusClientSecret string `env:"GENIUS_API_SECRET"`
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(".env"); err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			slog.Info("No .env file found, defaulting to environment variables")
		} else {
			return nil, fmt.Errorf("failed to load .env file: %w", err)
		}
	}

	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return cfg, nil
}
