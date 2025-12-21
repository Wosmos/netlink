package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go-to-do/auth"
	"go-to-do/config"
	"go-to-do/handlers"
	"go-to-do/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Load .env file
	config.LoadEnv()

	// PostgreSQL connection string from .env or environment
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL not set. Add it to .env file or set as environment variable")
	}

	// Connection pool config
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatal("Failed to parse config:", err)
	}

	// Pool settings for high concurrency
	poolConfig.MaxConns = 100
	poolConfig.MinConns = 10
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	fmt.Println("✅ Connected to PostgreSQL")

	// Initialize repositories
	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	taskRepo := repository.NewTaskRepository(pool)

	// Initialize schemas
	ctx := context.Background()
	if err := userRepo.InitSchema(ctx); err != nil {
		log.Fatal("Failed to init user schema:", err)
	}
	if err := sessionRepo.InitSchema(ctx); err != nil {
		log.Fatal("Failed to init session schema:", err)
	}
	if err := taskRepo.InitSchema(ctx); err != nil {
		log.Fatal("Failed to init task schema:", err)
	}
	fmt.Println("✅ Database schema initialized")

	// Initialize services
	authService := auth.NewAuthService(userRepo, sessionRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	taskHandler := handlers.NewTaskHandler(taskRepo, authService)

	// Routes
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	http.HandleFunc("/login", authHandler.Login)
	http.HandleFunc("/register", authHandler.Register)
	http.HandleFunc("/logout", authHandler.Logout)
	http.HandleFunc("/", taskHandler.Home)
	http.HandleFunc("/add", taskHandler.Add)
	http.HandleFunc("/toggle", taskHandler.Toggle)
	http.HandleFunc("/delete", taskHandler.Delete)

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Println("🚀 Server starting on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
