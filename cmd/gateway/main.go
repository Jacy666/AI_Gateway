package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Jacy666/AI_Gateway/internal/config"
	"github.com/Jacy666/AI_Gateway/internal/grpc"
	"github.com/Jacy666/AI_Gateway/internal/middleware"
	"github.com/Jacy666/AI_Gateway/internal/redis"
	"github.com/Jacy666/AI_Gateway/internal/task"
	"github.com/Jacy666/AI_Gateway/pkg/logger"
	"github.com/Jacy666/AI_Gateway/pkg/response"
	"github.com/gin-gonic/gin"
)

var (
	taskQueue  *task.TaskQueue
	redisStore *redis.Storage
	grpcClient *grpc.Client
	appConfig  *config.Config
)

func main() {
	// Load configuration
	appConfig = config.Load()

	// Set Gin mode
	gin.SetMode(appConfig.Server.Mode)

	// Initialize Redis storage
	var err error
	redisStore, err = redis.NewStorage(appConfig)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to initialize Redis storage: %v", err)
		logger.InfoLogger.Println("Continuing without Redis - tasks will not persist")
		// Create a mock storage for development
		redisStore = nil
	} else {
		logger.InfoLogger.Println("Redis storage initialized successfully")
	}

	// Initialize gRPC client
	grpcClient, err = grpc.NewClient(appConfig)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to initialize gRPC client: %v", err)
		logger.InfoLogger.Println("Continuing with mock gRPC client")
	} else {
		logger.InfoLogger.Println("gRPC client initialized successfully")
	}

	// Initialize task queue
	if redisStore != nil {
		taskQueue = task.NewTaskQueue(appConfig, redisStore, grpcClient)
	} else {
		logger.InfoLogger.Println("Task queue not initialized due to missing Redis storage")
	}

	// Setup router
	router := setupRouter()

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + appConfig.Server.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.InfoLogger.Printf("Starting server on port %s", appConfig.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.ErrorLogger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.InfoLogger.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.ErrorLogger.Fatal("Server forced to shutdown:", err)
	}

	// Cleanup resources
	if taskQueue != nil {
		taskQueue.Shutdown()
	}
	if grpcClient != nil {
		grpcClient.Close()
	}
	if redisStore != nil {
		redisStore.Close()
	}

	logger.InfoLogger.Println("Server exited")
}

func setupRouter() *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(middleware.Logger())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RateLimiter(appConfig))
	router.Use(gin.Recovery())

	// Health check endpoint (no auth required)
	router.GET("/health", healthHandler)

	// Auth endpoints (no JWT required)
	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/login", loginHandler)
		auth.POST("/register", registerHandler)
	}

	// Protected API endpoints
	api := router.Group("/api/v1")
	api.Use(middleware.JWTAuth(appConfig))
	{
		// Task endpoints
		api.POST("/tasks", submitTaskHandler)
		api.GET("/tasks/:task_id", getTaskHandler)

		// Admin endpoints
		api.GET("/status", statusHandler)
	}

	return router
}

// healthHandler handles health check requests
func healthHandler(c *gin.Context) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
	}

	// Check Redis health
	if redisStore != nil {
		if err := redisStore.Health(); err != nil {
			health["redis"] = "unhealthy: " + err.Error()
		} else {
			health["redis"] = "healthy"
		}
	} else {
		health["redis"] = "not configured"
	}

	// Check gRPC health
	if grpcClient != nil {
		if err := grpcClient.Health(); err != nil {
			health["grpc"] = "unhealthy: " + err.Error()
		} else {
			health["grpc"] = "healthy"
		}
	} else {
		health["grpc"] = "not configured"
	}

	response.Success(c, health)
}

// loginHandler handles user login
func loginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// TODO: Implement actual authentication logic
	// For now, accept any non-empty credentials
	if req.Username == "" || req.Password == "" {
		response.Error(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(appConfig, "user-"+req.Username, req.Username)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	response.Success(c, gin.H{
		"token": token,
		"user": gin.H{
			"username": req.Username,
		},
	})
}

// registerHandler handles user registration
func registerHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// TODO: Implement actual user registration logic
	// For now, just return success

	response.Created(c, gin.H{
		"username": req.Username,
		"message":  "User registered successfully",
	})
}

// submitTaskHandler handles task submission
func submitTaskHandler(c *gin.Context) {
	if taskQueue == nil {
		response.Error(c, http.StatusServiceUnavailable, "Task queue is not available")
		return
	}

	var req struct {
		Type    string                 `json:"type" binding:"required"`
		Payload map[string]interface{} `json:"payload" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Get user info from context
	userID, _ := c.Get("user_id")

	// Add user info to payload
	if req.Payload == nil {
		req.Payload = make(map[string]interface{})
	}
	req.Payload["user_id"] = userID

	// Submit task to queue
	taskID, err := taskQueue.SubmitTask(req.Type, req.Payload)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to submit task: %v", err))
		return
	}

	response.Created(c, gin.H{
		"task_id": taskID,
		"message": "Task submitted successfully",
	})
}

// getTaskHandler handles task status query
func getTaskHandler(c *gin.Context) {
	if taskQueue == nil {
		response.Error(c, http.StatusServiceUnavailable, "Task queue is not available")
		return
	}

	taskID := c.Param("task_id")
	if taskID == "" {
		response.Error(c, http.StatusBadRequest, "Task ID is required")
		return
	}

	// Get task from storage
	taskResult, err := taskQueue.GetTask(taskID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Task not found")
		return
	}

	response.Success(c, taskResult)
}

// statusHandler returns the gateway status
func statusHandler(c *gin.Context) {
	status := map[string]interface{}{
		"status":    "running",
		"timestamp": time.Now().Unix(),
		"config": map[string]interface{}{
			"num_workers":    appConfig.Worker.NumWorkers,
			"buffer_size":    appConfig.Worker.BufferSize,
			"rate_limit_rps": appConfig.RateLimit.RequestsPerSecond,
		},
	}

	response.Success(c, status)
}
