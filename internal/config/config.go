package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration
type Config struct {
	Server    ServerConfig
	Redis     RedisConfig
	JWT       JWTConfig
	GRPC      GRPCConfig
	RateLimit RateLimitConfig
	Worker    WorkerConfig
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Port string
	Mode string
}

// RedisConfig contains Redis connection configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// JWTConfig contains JWT authentication configuration
type JWTConfig struct {
	Secret     string
	ExpireTime time.Duration
}

// GRPCConfig contains gRPC client configuration
type GRPCConfig struct {
	ModelServiceAddr string
	Timeout          time.Duration
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
}

// WorkerConfig contains worker pool configuration
type WorkerConfig struct {
	NumWorkers int
	BufferSize int
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key"),
			ExpireTime: time.Duration(getEnvAsInt("JWT_EXPIRE_HOURS", 24)) * time.Hour,
		},
		GRPC: GRPCConfig{
			ModelServiceAddr: getEnv("GRPC_MODEL_SERVICE", "localhost:50051"),
			Timeout:          time.Duration(getEnvAsInt("GRPC_TIMEOUT_SECONDS", 60)) * time.Second,
		},
		RateLimit: RateLimitConfig{
			RequestsPerSecond: getEnvAsFloat("RATE_LIMIT_RPS", 100.0),
			BurstSize:         getEnvAsInt("RATE_LIMIT_BURST", 200),
		},
		Worker: WorkerConfig{
			NumWorkers: getEnvAsInt("NUM_WORKERS", 10),
			BufferSize: getEnvAsInt("TASK_BUFFER_SIZE", 1000),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultValue
}
