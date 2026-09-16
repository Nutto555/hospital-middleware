// Package config reads the service configuration from the environment.
package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds everything the service needs to start.
type Config struct {
	Port             string
	DatabaseURL      string
	JWTSecret        string
	JWTTTL           time.Duration
	HospitalABaseURL string
	HISTimeout       time.Duration
}

// Load reads the configuration from the environment. DATABASE_URL and JWT_SECRET are required.
func Load() (Config, error) {
	cfg := Config{
		Port:             getenv("PORT", "8080"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		HospitalABaseURL: getenv("HOSPITAL_A_BASE_URL", "https://hospital-a.api.co.th"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}
	var err error
	if cfg.JWTTTL, err = time.ParseDuration(getenv("JWT_TTL", "24h")); err != nil {
		return Config{}, fmt.Errorf("JWT_TTL: %w", err)
	}
	if cfg.HISTimeout, err = time.ParseDuration(getenv("HIS_TIMEOUT", "5s")); err != nil {
		return Config{}, fmt.Errorf("HIS_TIMEOUT: %w", err)
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
