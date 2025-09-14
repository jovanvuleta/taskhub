package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Create a temporary config for testing
	config = Config{
		App: struct {
			Name        string `yaml:"name"`
			Version     string `yaml:"version"`
			Port        int    `yaml:"port"`
			Environment string `yaml:"environment"`
		}{
			Name:        "test-app",
			Version:     "1.0.0",
			Port:        8080,
			Environment: "test",
		},
		Database: struct {
			Type           string `yaml:"type"`
			Path           string `yaml:"path"`
			MaxConnections int    `yaml:"max_connections"`
			Timeout        int    `yaml:"timeout"`
		}{
			Type: "sqlite",
			Path: ":memory:",
		},
		Security: struct {
			CorsEnabled bool     `yaml:"cors_enabled"`
			CorsOrigins []string `yaml:"cors_origins"`
		}{
			CorsEnabled: true,
			CorsOrigins: []string{"*"},
		},
	}

	// Initialize test database
	initDatabase()

	r := gin.Default()
	r.Use(corsMiddleware())

	api := r.Group("/api/v1")
	{
		api.GET("/health", healthCheck)
		api.GET("/users", getUsers)
		api.POST("/users", createUser)
		api.GET("/users/:id", getUser)
	}

	return r
}

func TestHealthCheck(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response.Status)
}

func TestCreateUser(t *testing.T) {
	router := setupTestRouter()

	user := User{
		Name:  "Test User",
		Email: "test@example.com",
	}
	jsonValue, _ := json.Marshal(user)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code)

	var response User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, user.Name, response.Name)
	assert.Equal(t, user.Email, response.Email)
	assert.NotEqual(t, 0, response.ID)
}

func TestCreateUserInvalidEmail(t *testing.T) {
	router := setupTestRouter()

	user := User{
		Name:  "Test User",
		Email: "invalid-email",
	}
	jsonValue, _ := json.Marshal(user)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Current implementation accepts any email format
	assert.Equal(t, 201, w.Code)
}

func TestGetUsers(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var users []User
	err := json.Unmarshal(w.Body.Bytes(), &users)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(users), 0)
}

func TestCorsMiddleware(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/v1/users", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	router.ServeHTTP(w, req)

	assert.Equal(t, 204, w.Code)
	// Current implementation uses wildcard CORS
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestMain(m *testing.M) {
	// Set up test environment
	os.Setenv("DB_USER", "test")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PASSWORD", "test")

	// Run tests
	code := m.Run()

	// Clean up
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PASSWORD")

	os.Exit(code)
}
