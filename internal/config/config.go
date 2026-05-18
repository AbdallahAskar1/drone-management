package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	JWTTTL      time.Duration
	AvgSpeedMS  float64
	LogLevel    string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:        getenv("PORT", "1323"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		LogLevel:    getenv("LOG_LEVEL", "info"),
	}

	ttlStr := getenv("JWT_TTL", "24h")
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		return nil, err
	}
	cfg.JWTTTL = ttl

	speedStr := getenv("AVG_SPEED_MS", "10")
	speed, err := strconv.ParseFloat(speedStr, 64)
	if err != nil {
		return nil, err
	}
	if speed <= 0 {
		return nil, errors.New("AVG_SPEED_MS must be > 0")
	}
	cfg.AvgSpeedMS = speed

	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}
	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
