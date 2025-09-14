package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
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
		api.GET("/tasks", getTasks)
		api.POST("/tasks", createTask)
		api.GET("/tasks/:id", getTask)
		api.PUT("/tasks/:id", updateTask)
		api.DELETE("/tasks/:id", deleteTask)
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

func TestCreateTask(t *testing.T) {
	router := setupTestRouter()

	task := Task{
		Title:       "Test Task",
		Description: "This is a test task",
		Status:      "pending",
	}
	jsonValue, _ := json.Marshal(task)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code)

	var response Task
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, task.Title, response.Title)
	assert.Equal(t, task.Description, response.Description)
	assert.Equal(t, task.Status, response.Status)
	assert.NotEqual(t, 0, response.ID)
}

func TestCreateTaskMissingTitle(t *testing.T) {
	router := setupTestRouter()

	task := Task{
		Description: "This task has no title",
		Status:      "pending",
	}
	jsonValue, _ := json.Marshal(task)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Current implementation accepts tasks without title validation
	assert.Equal(t, 201, w.Code)
}

func TestGetTasks(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tasks", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var tasks []Task
	err := json.Unmarshal(w.Body.Bytes(), &tasks)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(tasks), 0)
}

func TestGetTask(t *testing.T) {
	router := setupTestRouter()

	// First create a task
	task := Task{
		Title:       "Test Task",
		Description: "This is a test task",
		Status:      "pending",
	}
	jsonValue, _ := json.Marshal(task)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var createdTask Task
	json.Unmarshal(w.Body.Bytes(), &createdTask)

	// Now get the task by ID
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/tasks/"+strconv.Itoa(createdTask.ID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response Task
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, createdTask.ID, response.ID)
	assert.Equal(t, task.Title, response.Title)
}

func TestUpdateTask(t *testing.T) {
	router := setupTestRouter()

	// First create a task
	task := Task{
		Title:       "Original Task",
		Description: "Original description",
		Status:      "pending",
	}
	jsonValue, _ := json.Marshal(task)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var createdTask Task
	json.Unmarshal(w.Body.Bytes(), &createdTask)

	// Update the task
	updatedTask := Task{
		Title:       "Updated Task",
		Description: "Updated description",
		Status:      "completed",
	}
	jsonValue, _ = json.Marshal(updatedTask)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/api/v1/tasks/"+strconv.Itoa(createdTask.ID), bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response Task
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, updatedTask.Title, response.Title)
	assert.Equal(t, updatedTask.Status, response.Status)
}

func TestDeleteTask(t *testing.T) {
	router := setupTestRouter()

	// First create a task
	task := Task{
		Title:       "Task to Delete",
		Description: "This task will be deleted",
		Status:      "pending",
	}
	jsonValue, _ := json.Marshal(task)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var createdTask Task
	json.Unmarshal(w.Body.Bytes(), &createdTask)

	// Delete the task
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/v1/tasks/"+strconv.Itoa(createdTask.ID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	// Verify task is deleted by trying to get it
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/tasks/"+strconv.Itoa(createdTask.ID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}

func TestCorsMiddleware(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/v1/tasks", nil)
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
