package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	APIAddr             string
	PostgresHost        string
	PostgresPort        string
	PostgresDB          string
	PostgresAppUser     string
	PostgresAppPassword string
	AuthJWTSecret       string
	CORSOrigin          string
}

func Load() (Config, error) {
	var missing []string
	cfg := Config{
		APIAddr:             require("API_ADDR", &missing),
		PostgresHost:        require("POSTGRES_HOST", &missing),
		PostgresPort:        require("POSTGRES_PORT", &missing),
		PostgresDB:          require("POSTGRES_DB", &missing),
		PostgresAppUser:     require("POSTGRES_APP_USER", &missing),
		PostgresAppPassword: require("POSTGRES_APP_PASSWORD", &missing),
		AuthJWTSecret:       require("AUTH_JWT_SECRET", &missing),
		CORSOrigin:          os.Getenv("CORS_ORIGIN"),
	}
	if cfg.CORSOrigin == "" {
		cfg.CORSOrigin = "http://localhost:3000"
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required env: %s", strings.Join(missing, ", "))
	}
	if len(cfg.AuthJWTSecret) < 32 {
		return Config{}, fmt.Errorf("missing required env: AUTH_JWT_SECRET (must be at least 32 characters)")
	}
	return cfg, nil
}

func require(key string, missing *[]string) string {
	v := os.Getenv(key)
	if v == "" {
		*missing = append(*missing, key)
	}
	return v
}

func (c Config) AppDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.PostgresAppUser,
		c.PostgresAppPassword,
		c.PostgresHost,
		c.PostgresPort,
		c.PostgresDB,
	)
}
