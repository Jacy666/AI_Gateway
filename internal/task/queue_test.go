package task

import (
	"context"
	"testing"
	"time"

	"github.com/Jacy666/AI_Gateway/internal/config"
)

// MockStorage implements TaskStorage for testing
type MockStorage struct {
	tasks map[string]*Task
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		tasks: make(map[string]*Task),
	}
}

func (m *MockStorage) SaveTask(task *Task) error {
	m.tasks[task.ID] = task
	return nil
}

func (m *MockStorage) GetTask(taskID string) (*Task, error) {
	if task, ok := m.tasks[taskID]; ok {
		return task, nil
	}
	return nil, nil
}

func (m *MockStorage) UpdateTask(task *Task) error {
	m.tasks[task.ID] = task
	return nil
}

// MockGRPCClient implements GRPCClient for testing
type MockGRPCClient struct{}

func (m *MockGRPCClient) CallModelService(ctx context.Context, taskType string, payload map[string]interface{}) (interface{}, error) {
	// Simulate quick processing
	time.Sleep(10 * time.Millisecond)
	return map[string]interface{}{
		"result": "test result",
		"status": "success",
	}, nil
}

func TestTaskQueueSubmit(t *testing.T) {
	cfg := &config.Config{
		Worker: config.WorkerConfig{
			NumWorkers: 2,
			BufferSize: 10,
		},
		GRPC: config.GRPCConfig{
			Timeout: 30 * time.Second,
		},
	}
	
	storage := NewMockStorage()
	grpcClient := &MockGRPCClient{}
	
	queue := NewTaskQueue(cfg, storage, grpcClient)
	defer queue.Shutdown()
	
	// Submit a task
	taskID, err := queue.SubmitTask("test_task", map[string]interface{}{
		"key": "value",
	})
	
	if err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}
	
	if taskID == "" {
		t.Error("Expected non-empty task ID")
	}
	
	// Wait for task to be processed
	time.Sleep(100 * time.Millisecond)
	
	// Get task
	task, err := queue.GetTask(taskID)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	
	if task == nil {
		t.Fatal("Expected task to exist")
	}
	
	if task.Status != TaskStatusCompleted && task.Status != TaskStatusProcessing {
		t.Errorf("Expected task to be completed or processing, got %s", task.Status)
	}
}

func TestTaskStatus(t *testing.T) {
	task := &Task{
		ID:        "test-123",
		Type:      "test",
		Payload:   map[string]interface{}{"data": "test"},
		Status:    TaskStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	if task.Status != TaskStatusPending {
		t.Errorf("Expected status pending, got %s", task.Status)
	}
	
	// Test marshal/unmarshal
	data, err := MarshalTask(task)
	if err != nil {
		t.Fatalf("Failed to marshal task: %v", err)
	}
	
	unmarshaled, err := UnmarshalTask(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal task: %v", err)
	}
	
	if unmarshaled.ID != task.ID {
		t.Errorf("Expected ID %s, got %s", task.ID, unmarshaled.ID)
	}
	
	if unmarshaled.Status != task.Status {
		t.Errorf("Expected status %s, got %s", task.Status, unmarshaled.Status)
	}
}
