package task

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Jacy666/AI_Gateway/internal/config"
	"github.com/Jacy666/AI_Gateway/pkg/logger"
	"github.com/google/uuid"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

// Task represents an AI processing task
type Task struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Status    TaskStatus             `json:"status"`
	Result    interface{}            `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// TaskQueue manages the task queue with producer-consumer pattern
type TaskQueue struct {
	taskChan   chan *Task
	config     *config.Config
	storage    TaskStorage
	grpcClient GRPCClient
	ctx        context.Context
	cancel     context.CancelFunc
}

// TaskStorage interface for storing task status and results
type TaskStorage interface {
	SaveTask(task *Task) error
	GetTask(taskID string) (*Task, error)
	UpdateTask(task *Task) error
}

// GRPCClient interface for calling downstream Python model service
type GRPCClient interface {
	CallModelService(ctx context.Context, taskType string, payload map[string]interface{}) (interface{}, error)
}

// NewTaskQueue creates a new task queue
func NewTaskQueue(cfg *config.Config, storage TaskStorage, grpcClient GRPCClient) *TaskQueue {
	ctx, cancel := context.WithCancel(context.Background())

	tq := &TaskQueue{
		taskChan:   make(chan *Task, cfg.Worker.BufferSize),
		config:     cfg,
		storage:    storage,
		grpcClient: grpcClient,
		ctx:        ctx,
		cancel:     cancel,
	}

	// Start worker pool
	for i := 0; i < cfg.Worker.NumWorkers; i++ {
		go tq.worker(i)
	}

	logger.InfoLogger.Printf("Task queue initialized with %d workers and buffer size %d",
		cfg.Worker.NumWorkers, cfg.Worker.BufferSize)

	return tq
}

// SubmitTask submits a new task to the queue (Producer)
func (tq *TaskQueue) SubmitTask(taskType string, payload map[string]interface{}) (string, error) {
	task := &Task{
		ID:        uuid.New().String(),
		Type:      taskType,
		Payload:   payload,
		Status:    TaskStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save task to storage
	if err := tq.storage.SaveTask(task); err != nil {
		logger.ErrorLogger.Printf("Failed to save task: %v", err)
		return "", fmt.Errorf("failed to save task: %w", err)
	}

	// Push task to channel (non-blocking)
	select {
	case tq.taskChan <- task:
		logger.InfoLogger.Printf("Task submitted: %s", task.ID)
		return task.ID, nil
	default:
		// Channel is full, update task status to failed
		task.Status = TaskStatusFailed
		task.Error = "task queue is full"
		task.UpdatedAt = time.Now()
		_ = tq.storage.UpdateTask(task)
		return "", fmt.Errorf("task queue is full")
	}
}

// GetTask retrieves a task by ID
func (tq *TaskQueue) GetTask(taskID string) (*Task, error) {
	return tq.storage.GetTask(taskID)
}

// worker processes tasks from the queue (Consumer)
func (tq *TaskQueue) worker(workerID int) {
	logger.InfoLogger.Printf("Worker %d started", workerID)

	for {
		select {
		case <-tq.ctx.Done():
			logger.InfoLogger.Printf("Worker %d stopped", workerID)
			return
		case task := <-tq.taskChan:
			tq.processTask(workerID, task)
		}
	}
}

// processTask processes a single task
func (tq *TaskQueue) processTask(workerID int, task *Task) {
	logger.InfoLogger.Printf("Worker %d processing task: %s", workerID, task.ID)

	// Update task status to processing
	task.Status = TaskStatusProcessing
	task.UpdatedAt = time.Now()
	if err := tq.storage.UpdateTask(task); err != nil {
		logger.ErrorLogger.Printf("Failed to update task status: %v", err)
	}

	// Call gRPC service with timeout
	ctx, cancel := context.WithTimeout(tq.ctx, tq.config.GRPC.Timeout)
	defer cancel()

	result, err := tq.grpcClient.CallModelService(ctx, task.Type, task.Payload)

	// Update task with result or error
	if err != nil {
		task.Status = TaskStatusFailed
		task.Error = err.Error()
		logger.ErrorLogger.Printf("Worker %d failed to process task %s: %v", workerID, task.ID, err)
	} else {
		task.Status = TaskStatusCompleted
		task.Result = result
		logger.InfoLogger.Printf("Worker %d completed task: %s", workerID, task.ID)
	}

	task.UpdatedAt = time.Now()
	if err := tq.storage.UpdateTask(task); err != nil {
		logger.ErrorLogger.Printf("Failed to update task result: %v", err)
	}
}

// Shutdown gracefully shuts down the task queue
func (tq *TaskQueue) Shutdown() {
	logger.InfoLogger.Println("Shutting down task queue...")
	tq.cancel()
	close(tq.taskChan)
}

// MarshalTask converts a task to JSON
func MarshalTask(task *Task) ([]byte, error) {
	return json.Marshal(task)
}

// UnmarshalTask converts JSON to a task
func UnmarshalTask(data []byte) (*Task, error) {
	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, err
	}
	return &task, nil
}
