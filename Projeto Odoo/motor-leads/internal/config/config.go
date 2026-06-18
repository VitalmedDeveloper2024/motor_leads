package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr string
	DSN      string
	Workers  int
}

func Load() Config {
	return Config{
		HTTPAddr: env("HTTP_ADDR", ":8080"),
		DSN:      env("DATABASE_URL", "postgres://leads:leads@localhost:5432/leads?sslmode=disable"),
		Workers:  envInt("QUEUE_WORKERS", 8),
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}
