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

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
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
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE
	);`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		return err
	}

	insertSampleData := `
	INSERT OR IGNORE INTO users (name, email) VALUES 
		('John Doe', 'john@example.com'),
		('Jane Smith', 'jane@example.com'),
		('Bob Johnson', 'bob@example.com');`

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

func getUsers(c *gin.Context) {
	rows, err := db.Query("SELECT id, name, email FROM users")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}

func createUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", user.Name, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	user.ID = int(id)
	c.JSON(http.StatusCreated, user)
}

func getUser(c *gin.Context) {
	id := c.Param("id")
	var user User

	err := db.QueryRow("SELECT id, name, email FROM users WHERE id = ?", id).Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, user)
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
		api.GET("/users", getUsers)
		api.POST("/users", createUser)
		api.GET("/users/:id", getUser)
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
