package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v2"
)

type Config struct {
	App struct {
		Name        string `yaml:"name"`
		Version     string `yaml:"version"`
		Port        int    `yaml:"port"`
		Environment string `yaml:"environment"`
	} `yaml:"app"`
	Database struct {
		Type           string `yaml:"type"`
		Path           string `yaml:"path"`
		MaxConnections int    `yaml:"max_connections"`
		Timeout        int    `yaml:"timeout"`
	} `yaml:"database"`
	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`
	Security struct {
		CorsEnabled bool     `yaml:"cors_enabled"`
		CorsOrigins []string `yaml:"cors_origins"`
	} `yaml:"security"`
}

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}

var db *sql.DB
var config Config

func loadConfig(configPath string) error {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &config)
}

func initDatabase() error {
	dbUser := os.Getenv("DB_USER")
	dbHost := os.Getenv("DB_HOST")
	dbPassword := os.Getenv("DB_PASSWORD")

	log.Printf("Database config - User: %s, Host: %s, Password: %s",
		dbUser, dbHost, maskPassword(dbPassword))

	var err error
	db, err = sql.Open("sqlite3", config.Database.Path)
	if err != nil {
		return err
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		status TEXT DEFAULT 'pending',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		return err
	}

	insertSampleData := `
	INSERT OR IGNORE INTO tasks (title, description, status) VALUES 
		('Setup Development Environment', 'Install and configure development tools', 'completed'),
		('Create API Documentation', 'Document all API endpoints and responses', 'in_progress'),
		('Deploy to Production', 'Deploy application to production environment', 'pending');`

	_, err = db.Exec(insertSampleData)
	return err
}

func maskPassword(password string) string {
	if password == "" {
		return "not set"
	}
	return "***"
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func getTasks(c *gin.Context) {
	rows, err := db.Query("SELECT id, title, description, status, created_at FROM tasks ORDER BY id DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		tasks = append(tasks, task)
	}

	c.JSON(http.StatusOK, tasks)
}

func createTask(c *gin.Context) {
	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if task.Status == "" {
		task.Status = "pending"
	}

	result, err := db.Exec("INSERT INTO tasks (title, description, status) VALUES (?, ?, ?)", task.Title, task.Description, task.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	task.ID = int(id)

	// Get the created_at timestamp
	err = db.QueryRow("SELECT created_at FROM tasks WHERE id = ?", task.ID).Scan(&task.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

func getTask(c *gin.Context) {
	id := c.Param("id")
	var task Task

	err := db.QueryRow("SELECT id, title, description, status, created_at FROM tasks WHERE id = ?", id).Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, task)
}

func updateTask(c *gin.Context) {
	id := c.Param("id")
	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := db.Exec("UPDATE tasks SET title = ?, description = ?, status = ? WHERE id = ?", task.Title, task.Description, task.Status, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Get the updated task
	err = db.QueryRow("SELECT id, title, description, status, created_at FROM tasks WHERE id = ?", id).Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

func deleteTask(c *gin.Context) {
	id := c.Param("id")

	result, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

func healthCheck(c *gin.Context) {
	response := HealthResponse{
		Status:    "healthy",
		Version:   config.App.Version,
		Timestamp: fmt.Sprintf("%d", c.Request.Context().Value("timestamp")),
	}
	c.JSON(http.StatusOK, response)
}

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config.yaml"
	}

	if err := loadConfig(configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := initDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	if config.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

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

	port := config.App.Port
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	log.Printf("Starting %s v%s on port %d", config.App.Name, config.App.Version, port)
	log.Fatal(r.Run(fmt.Sprintf(":%d", port)))
}