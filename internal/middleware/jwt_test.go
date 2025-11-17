package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Jacy666/AI_Gateway/internal/config"
	"github.com/gin-gonic/gin"
)

func TestGenerateAndValidateToken(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret",
			ExpireTime: 24 * time.Hour,
		},
	}
	
	// Generate token
	token, err := GenerateToken(cfg, "user123", "testuser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	
	if token == "" {
		t.Error("Expected non-empty token")
	}
	
	// Setup test server with JWT middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(JWTAuth(cfg))
	router.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")
		c.JSON(http.StatusOK, gin.H{
			"user_id":  userID,
			"username": username,
		})
	})
	
	// Test with valid token
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestJWTAuthMissingToken(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret",
			ExpireTime: 24 * time.Hour,
		},
	}
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(JWTAuth(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	
	// Test without token
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}
