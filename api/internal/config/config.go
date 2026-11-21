package config

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config holds every setting the API reads from the environment.
type Config struct {
	AppEnv      string `env:"APP_ENV" envDefault:"development"`
	HTTPAddr    string `env:"HTTP_ADDR"`
	Port        string `env:"PORT"` // injected by Render in production
	DatabaseURL string `env:"DATABASE_URL,required,notEmpty"`
	RedisURL    string `env:"REDIS_URL,required,notEmpty"`
}

// Load reads an optional .env file, then parses environment variables into Config.
func Load() (Config, error) {
	// .env exists only for native local dev, so a missing file is fine.
	// godotenv never overrides variables that are already set,
	// so values from Docker or Render always win over the file.
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("parse env: %w", err)
	}

	if cfg.HTTPAddr == "" {
		if cfg.Port != "" {
			cfg.HTTPAddr = ":" + cfg.Port
		} else {
			cfg.HTTPAddr = "localhost:8080"
		}
	}

	return cfg, nil
}

func (c Config) IsProduction() bool {
	return c.AppEnv == "production"
}