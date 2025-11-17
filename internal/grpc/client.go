package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Jacy666/AI_Gateway/internal/config"
	"github.com/Jacy666/AI_Gateway/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client implements the GRPCClient interface
type Client struct {
	conn   *grpc.ClientConn
	config *config.Config
}

// NewClient creates a new gRPC client
func NewClient(cfg *config.Config) (*Client, error) {
	// Connect to gRPC server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, cfg.GRPC.ModelServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		// Log warning but don't fail - the service might not be ready yet
		logger.ErrorLogger.Printf("Warning: Failed to connect to gRPC service at %s: %v", cfg.GRPC.ModelServiceAddr, err)
		// Return a client with nil connection for graceful degradation
		return &Client{
			conn:   nil,
			config: cfg,
		}, nil
	}

	logger.InfoLogger.Printf("Connected to gRPC service at %s", cfg.GRPC.ModelServiceAddr)

	return &Client{
		conn:   conn,
		config: cfg,
	}, nil
}

// CallModelService calls the downstream Python model service via gRPC
// For now, this is a mock implementation that simulates calling a model service
func (c *Client) CallModelService(ctx context.Context, taskType string, payload map[string]interface{}) (interface{}, error) {
	// Check if connection is available
	if c.conn == nil {
		// Mock response for demonstration when gRPC service is not available
		logger.InfoLogger.Printf("Simulating model service call for task type: %s", taskType)

		// Simulate processing time
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled")
		case <-time.After(2 * time.Second):
			// Continue with mock response
		}

		mockResult := map[string]interface{}{
			"task_type":    taskType,
			"status":       "completed",
			"model":        "mock-model-v1",
			"result":       "This is a mock result from the AI model service",
			"confidence":   0.95,
			"processed_at": time.Now().Format(time.RFC3339),
		}

		return mockResult, nil
	}

	// TODO: Implement actual gRPC call using generated protobuf stubs
	// This would involve:
	// 1. Creating protobuf definitions for the model service API
	// 2. Generating Go code from the proto files
	// 3. Using the generated client to make the actual RPC call

	// For now, return a placeholder implementation
	logger.InfoLogger.Printf("Calling model service for task type: %s", taskType)

	// Simulate processing time
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled")
	case <-time.After(2 * time.Second):
		// Continue
	}

	result := map[string]interface{}{
		"task_type": taskType,
		"payload":   payload,
		"result":    "Processed by model service",
		"timestamp": time.Now().Unix(),
	}

	return result, nil
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Health checks if the gRPC connection is healthy
func (c *Client) Health() error {
	if c.conn == nil {
		return fmt.Errorf("gRPC connection not established")
	}
	return nil
}

// MockClient is a mock implementation for testing
type MockClient struct{}

// NewMockClient creates a new mock gRPC client
func NewMockClient() *MockClient {
	return &MockClient{}
}

// CallModelService implements a mock model service call
func (m *MockClient) CallModelService(ctx context.Context, taskType string, payload map[string]interface{}) (interface{}, error) {
	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	payloadJSON, _ := json.Marshal(payload)
	return map[string]interface{}{
		"task_type": taskType,
		"payload":   string(payloadJSON),
		"result":    "Mock result",
		"timestamp": time.Now().Unix(),
	}, nil
}
