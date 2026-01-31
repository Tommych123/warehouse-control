package config

import (
	"errors"
	"os"
)

type Config struct {
	Addr      string
	DBDSN     string
	JWTSecret string
	Env       string
}

func Load() (Config, error) {
	cfg := Config{
		Addr:      getEnv("ADDR", ":8080"),
		DBDSN:     getEnv("DB_DSN", ""),
		JWTSecret: getEnv("JWT_SECRET", ""),
		Env:       getEnv("ENV", "dev"),
	}

	if cfg.JWTSecret == "" {
		return Config{}, errors.New("JWT_SECRET is required")
	}
	if cfg.DBDSN == "" {
		return Config{}, errors.New("DB_DSN is required")
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
