package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Set environment variables
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("REDIS_HOST", "testhost")
	os.Setenv("NUM_WORKERS", "20")
	
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("NUM_WORKERS")
	}()
	
	cfg := Load()
	
	if cfg.Server.Port != "9090" {
		t.Errorf("Expected port 9090, got %s", cfg.Server.Port)
	}
	
	if cfg.Redis.Host != "testhost" {
		t.Errorf("Expected redis host testhost, got %s", cfg.Redis.Host)
	}
	
	if cfg.Worker.NumWorkers != 20 {
		t.Errorf("Expected 20 workers, got %d", cfg.Worker.NumWorkers)
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg := Load()
	
	if cfg.Server.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", cfg.Server.Port)
	}
	
	if cfg.Worker.NumWorkers != 10 {
		t.Errorf("Expected default 10 workers, got %d", cfg.Worker.NumWorkers)
	}
	
	if cfg.JWT.ExpireTime != 24*time.Hour {
		t.Errorf("Expected default JWT expire time 24h, got %v", cfg.JWT.ExpireTime)
	}
}
