package config_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Nutto555/hospital-middleware/internal/config"
)

func setAll(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/db")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("PORT", "")
	t.Setenv("JWT_TTL", "")
	t.Setenv("HIS_TIMEOUT", "")
	t.Setenv("HOSPITAL_A_BASE_URL", "")
}

func TestLoadDefaults(t *testing.T) {
	setAll(t)
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "8080" || cfg.JWTTTL != 24*time.Hour || cfg.HISTimeout != 5*time.Second || cfg.HospitalABaseURL != "https://hospital-a.api.co.th" {
		t.Errorf("unexpected defaults: %+v", cfg)
	}
}

func TestLoadOverrides(t *testing.T) {
	setAll(t)
	t.Setenv("PORT", "9000")
	t.Setenv("JWT_TTL", "1h")
	t.Setenv("HIS_TIMEOUT", "250ms")
	t.Setenv("HOSPITAL_A_BASE_URL", "http://hospital-a:8081")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "9000" || cfg.JWTTTL != time.Hour || cfg.HISTimeout != 250*time.Millisecond || cfg.HospitalABaseURL != "http://hospital-a:8081" {
		t.Errorf("overrides not applied: %+v", cfg)
	}
}

func TestLoadRequired(t *testing.T) {
	for _, key := range []string{"DATABASE_URL", "JWT_SECRET"} {
		t.Run(key, func(t *testing.T) {
			setAll(t)
			t.Setenv(key, "")
			_, err := config.Load()
			if err == nil || !strings.Contains(err.Error(), key) {
				t.Fatalf("want error naming %s, got %v", key, err)
			}
		})
	}
}

func TestLoadBadDuration(t *testing.T) {
	setAll(t)
	t.Setenv("JWT_TTL", "soon")
	if _, err := config.Load(); err == nil {
		t.Fatal("want error for bad JWT_TTL")
	}
}
