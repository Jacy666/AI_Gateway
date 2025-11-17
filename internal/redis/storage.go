package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/Jacy666/AI_Gateway/internal/config"
	"github.com/Jacy666/AI_Gateway/internal/task"
	"github.com/redis/go-redis/v9"
)

const (
	taskKeyPrefix = "task:"
	taskTTL       = 24 * time.Hour
)

// Storage implements TaskStorage interface using Redis
type Storage struct {
	client *redis.Client
	ctx    context.Context
}

// NewStorage creates a new Redis storage instance
func NewStorage(cfg *config.Config) (*Storage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Storage{
		client: client,
		ctx:    ctx,
	}, nil
}

// SaveTask saves a task to Redis
func (s *Storage) SaveTask(t *task.Task) error {
	data, err := task.MarshalTask(t)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	key := taskKeyPrefix + t.ID
	if err := s.client.Set(s.ctx, key, data, taskTTL).Err(); err != nil {
		return fmt.Errorf("failed to save task to Redis: %w", err)
	}

	return nil
}

// GetTask retrieves a task from Redis
func (s *Storage) GetTask(taskID string) (*task.Task, error) {
	key := taskKeyPrefix + taskID
	data, err := s.client.Get(s.ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task from Redis: %w", err)
	}

	t, err := task.UnmarshalTask(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return t, nil
}

// UpdateTask updates a task in Redis
func (s *Storage) UpdateTask(t *task.Task) error {
	return s.SaveTask(t)
}

// Close closes the Redis connection
func (s *Storage) Close() error {
	return s.client.Close()
}

// Health checks if Redis is healthy
func (s *Storage) Health() error {
	return s.client.Ping(s.ctx).Err()
}

// TaskResult represents the result for API response
type TaskResult struct {
	ID        string      `json:"id"`
	Status    string      `json:"status"`
	Result    interface{} `json:"result,omitempty"`
	Error     string      `json:"error,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// GetTaskResult retrieves a task result for API response
func (s *Storage) GetTaskResult(taskID string) (*TaskResult, error) {
	t, err := s.GetTask(taskID)
	if err != nil {
		return nil, err
	}

	return &TaskResult{
		ID:        t.ID,
		Status:    string(t.Status),
		Result:    t.Result,
		Error:     t.Error,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}, nil
}

// StoreToken stores a JWT token with expiration
func (s *Storage) StoreToken(userID, token string, expiration time.Duration) error {
	key := "token:" + userID
	return s.client.Set(s.ctx, key, token, expiration).Err()
}

// GetToken retrieves a JWT token
func (s *Storage) GetToken(userID string) (string, error) {
	key := "token:" + userID
	return s.client.Get(s.ctx, key).Result()
}

// DeleteToken deletes a JWT token
func (s *Storage) DeleteToken(userID string) error {
	key := "token:" + userID
	return s.client.Del(s.ctx, key).Err()
}
