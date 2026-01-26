package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"go-to-do/auth"
	"go-to-do/config"
	"go-to-do/handlers"
	"go-to-do/models"
	"go-to-do/repository"
	"go-to-do/websocket"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Printf("🔧 Using %d CPU cores\n", runtime.NumCPU())

	config.LoadEnv()

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL not set")
	}

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatal("Failed to parse config:", err)
	}

	poolConfig.MaxConns = 200
	poolConfig.MinConns = 20
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute
	poolConfig.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	fmt.Println("✅ Connected to PostgreSQL")

	// Initialize repositories
	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	taskRepo := repository.NewTaskRepository(pool)
	noteRepo := repository.NewNoteRepository(pool)
	chatRepo := repository.NewChatRepository(pool)

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
	if err := noteRepo.InitSchema(ctx); err != nil {
		log.Fatal("Failed to init note schema:", err)
	}
	if err := chatRepo.InitSchema(ctx); err != nil {
		log.Fatal("Failed to init chat schema:", err)
	}
	if err := chatRepo.InitReactionsSchema(ctx); err != nil {
		log.Fatal("Failed to init reactions schema:", err)
	}
	fmt.Println("✅ Database schema initialized")

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()
	fmt.Println("✅ WebSocket hub started")

	// Initialize services
	authService := auth.NewAuthService(userRepo, sessionRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	taskHandler := handlers.NewTaskHandler(taskRepo, authService)
	noteHandler := handlers.NewNoteHandler(noteRepo, authService)
	chatHandler := handlers.NewChatHandler(chatRepo, userRepo, authService, hub)

	// CORS middleware for API routes
	corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				origin = "http://localhost:3000"
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next(w, r)
		}
	}

	// Static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	// Auth routes (HTML pages)
	http.HandleFunc("/login", authHandler.Login)
	http.HandleFunc("/register", authHandler.Register)
	http.HandleFunc("/logout", authHandler.Logout)
	http.HandleFunc("/verify", authHandler.Verify)
	http.HandleFunc("/forgot-password", authHandler.ForgotPassword)
	http.HandleFunc("/reset-password", authHandler.ResetPassword)

	// Task routes (HTML pages)
	http.HandleFunc("/", taskHandler.Home)
	http.HandleFunc("/add", taskHandler.Add)
	http.HandleFunc("/toggle", taskHandler.Toggle)
	http.HandleFunc("/delete", taskHandler.Delete)

	// ============ API Routes ============

	// Auth API
	http.HandleFunc("/api/auth/me", corsMiddleware(authHandler.Me))
	http.HandleFunc("/api/auth/login", corsMiddleware(authHandler.APILogin))
	http.HandleFunc("/api/auth/register", corsMiddleware(authHandler.APIRegister))
	http.HandleFunc("/api/auth/logout", corsMiddleware(authHandler.APILogout))
	http.HandleFunc("/api/auth/forgot-password", corsMiddleware(authHandler.APIForgotPassword))
	http.HandleFunc("/api/auth/reset-password", corsMiddleware(authHandler.APIResetPassword))
	http.HandleFunc("/api/auth/verify", corsMiddleware(authHandler.APIVerify))
	http.HandleFunc("/api/test-email", corsMiddleware(authHandler.TestEmail))

	// Tasks API
	http.HandleFunc("/api/tasks", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			taskHandler.APIList(w, r)
		case "POST":
			taskHandler.APICreate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	http.HandleFunc("/api/tasks/toggle", corsMiddleware(taskHandler.APIToggle))
	http.HandleFunc("/api/tasks/delete", corsMiddleware(taskHandler.APIDelete))

	// Notes API
	http.HandleFunc("/api/notes", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			noteHandler.List(w, r)
		case "POST":
			noteHandler.Create(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	http.HandleFunc("/api/notes/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			noteHandler.Get(w, r)
		case "PUT":
			noteHandler.Update(w, r)
		case "DELETE":
			noteHandler.Delete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	http.HandleFunc("/api/notes/pin", corsMiddleware(noteHandler.TogglePin))

	// Chat API
	http.HandleFunc("/api/conversations", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			chatHandler.ListConversations(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	http.HandleFunc("/api/conversations/direct", corsMiddleware(chatHandler.CreateDirectChat))
	http.HandleFunc("/api/conversations/group", corsMiddleware(chatHandler.CreateGroup))
	http.HandleFunc("/api/conversations/messages", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			chatHandler.GetMessages(w, r)
		case "POST":
			chatHandler.SendMessage(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	http.HandleFunc("/api/conversations/messages/edit", corsMiddleware(chatHandler.EditMessage))
	http.HandleFunc("/api/conversations/read", corsMiddleware(chatHandler.MarkAsRead))
	http.HandleFunc("/api/conversations/typing", corsMiddleware(chatHandler.SendTyping))

	// Message actions
	http.HandleFunc("/api/messages/react", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			chatHandler.ReactToMessage(w, r)
		case "DELETE":
			chatHandler.RemoveReaction(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	http.HandleFunc("/api/messages/forward", corsMiddleware(chatHandler.ForwardMessage))
	http.HandleFunc("/api/messages/delete", corsMiddleware(chatHandler.DeleteMessage))
	http.HandleFunc("/api/conversations/delete", corsMiddleware(chatHandler.DeleteConversation))

	// Users API
	http.HandleFunc("/api/users/online", corsMiddleware(chatHandler.GetOnlineUsers))
	http.HandleFunc("/api/users/search", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		userID := chatHandler.RequireAuthPublic(w, r)
		if userID == 0 {
			return
		}
		query := r.URL.Query().Get("q")
		if query == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Query required"})
			return
		}
		users, err := userRepo.SearchUsers(query, 20)
		if err != nil {
			log.Printf("Search error: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Search failed"})
			return
		}
		if users == nil {
			users = []models.User{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": users})
	}))

	// WebSocket
	http.HandleFunc("/ws", chatHandler.HandleWebSocket)

	server := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	fmt.Println("🚀 Server starting on http://localhost:8080")
	fmt.Println("📡 WebSocket available at ws://localhost:8080/ws")
	log.Fatal(server.ListenAndServe())
}
