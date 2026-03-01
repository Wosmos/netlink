package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"netlink/auth"
	"netlink/config"
	"netlink/handlers"
	"netlink/middleware"
	"netlink/models"
	"netlink/repository"
	"netlink/websocket"

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

	// Periodic session cleanup — remove expired sessions every hour
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := sessionRepo.CleanExpired(); err != nil {
				log.Printf("Session cleanup error: %v", err)
			}
		}
	}()

	// Build allowed origins set from ALLOWED_ORIGINS env var
	allowedOrigins := map[string]bool{
		"http://localhost:3000": true,
		"http://localhost:8080": true,
	}
	if extra := os.Getenv("ALLOWED_ORIGINS"); extra != "" {
		for _, o := range strings.Split(extra, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				allowedOrigins[o] = true
			}
		}
	}

	// Initialize services
	authService := auth.NewAuthService(userRepo, sessionRepo)

	// Initialize rate limiter: 20 requests per minute for auth endpoints
	authLimiter := middleware.NewRateLimiter(20, 1*time.Minute)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	taskHandler := handlers.NewTaskHandler(taskRepo, authService)
	noteHandler := handlers.NewNoteHandler(noteRepo, authService)
	chatHandler := handlers.NewChatHandler(chatRepo, userRepo, authService, hub, allowedOrigins)
	voiceHandler := handlers.NewVoiceHandler(chatRepo, userRepo, authService)

	// CORS middleware for API routes
	corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if allowedOrigins[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
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

	// Body size limit for JSON endpoints (1 MB)
	jsonLimit := func(next http.HandlerFunc) http.HandlerFunc {
		return middleware.LimitBody(1<<20, next)
	}

	// Auth API (rate-limited to prevent brute force)
	http.HandleFunc("/api/auth/me", corsMiddleware(authHandler.Me))
	http.HandleFunc("/api/auth/login", corsMiddleware(authLimiter.Middleware(jsonLimit(authHandler.APILogin))))
	http.HandleFunc("/api/auth/register", corsMiddleware(authLimiter.Middleware(jsonLimit(authHandler.APIRegister))))
	http.HandleFunc("/api/auth/logout", corsMiddleware(authHandler.APILogout))
	http.HandleFunc("/api/auth/forgot-password", corsMiddleware(authLimiter.Middleware(jsonLimit(authHandler.APIForgotPassword))))
	http.HandleFunc("/api/auth/reset-password", corsMiddleware(authLimiter.Middleware(jsonLimit(authHandler.APIResetPassword))))
	http.HandleFunc("/api/auth/verify", corsMiddleware(authHandler.APIVerify))
	http.HandleFunc("/api/test-email", corsMiddleware(authLimiter.Middleware(jsonLimit(authHandler.TestEmail))))

	// Tasks API
	http.HandleFunc("/api/tasks", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			taskHandler.APIList(w, r)
		case "POST":
			jsonLimit(taskHandler.APICreate)(w, r)
		default:
			middleware.JSONMethodNotAllowed(w)
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
			jsonLimit(noteHandler.Create)(w, r)
		default:
			middleware.JSONMethodNotAllowed(w)
		}
	}))
	http.HandleFunc("/api/notes/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			noteHandler.Get(w, r)
		case "PUT":
			jsonLimit(noteHandler.Update)(w, r)
		case "DELETE":
			noteHandler.Delete(w, r)
		default:
			middleware.JSONMethodNotAllowed(w)
		}
	}))
	http.HandleFunc("/api/notes/pin", corsMiddleware(noteHandler.TogglePin))

	// Chat API
	http.HandleFunc("/api/conversations", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			chatHandler.ListConversations(w, r)
		} else {
			middleware.JSONMethodNotAllowed(w)
		}
	}))
	http.HandleFunc("/api/conversations/direct", corsMiddleware(jsonLimit(chatHandler.CreateDirectChat)))
	http.HandleFunc("/api/conversations/group", corsMiddleware(jsonLimit(chatHandler.CreateGroup)))
	http.HandleFunc("/api/conversations/messages", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			chatHandler.GetMessages(w, r)
		case "POST":
			jsonLimit(chatHandler.SendMessage)(w, r)
		default:
			middleware.JSONMethodNotAllowed(w)
		}
	}))
	http.HandleFunc("/api/conversations/messages/edit", corsMiddleware(jsonLimit(chatHandler.EditMessage)))
	http.HandleFunc("/api/conversations/read", corsMiddleware(chatHandler.MarkAsRead))
	http.HandleFunc("/api/conversations/typing", corsMiddleware(chatHandler.SendTyping))

	// Message actions
	http.HandleFunc("/api/messages/react", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			jsonLimit(chatHandler.ReactToMessage)(w, r)
		case "DELETE":
			chatHandler.RemoveReaction(w, r)
		default:
			middleware.JSONMethodNotAllowed(w)
		}
	}))
	http.HandleFunc("/api/messages/forward", corsMiddleware(jsonLimit(chatHandler.ForwardMessage)))
	http.HandleFunc("/api/messages/delete", corsMiddleware(chatHandler.DeleteMessage))
	http.HandleFunc("/api/conversations/delete", corsMiddleware(chatHandler.DeleteConversation))

	// Voice messages API
	http.HandleFunc("/api/voice/upload", corsMiddleware(voiceHandler.UploadVoice))
	http.HandleFunc("/api/voice/download", corsMiddleware(voiceHandler.DownloadVoice))
	http.HandleFunc("/api/voice/delete", corsMiddleware(voiceHandler.DeleteVoice))

	// Users API
	http.HandleFunc("/api/users/online", corsMiddleware(chatHandler.GetOnlineUsers))
	http.HandleFunc("/api/users/search", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		userID := chatHandler.RequireAuthPublic(w, r)
		if userID == 0 {
			return
		}
		query := r.URL.Query().Get("q")
		if query == "" {
			middleware.JSONError(w, "Query required", http.StatusBadRequest)
			return
		}
		users, err := userRepo.SearchUsers(query, 20)
		if err != nil {
			log.Printf("Search error: %v", err)
			middleware.JSONError(w, "Search failed", http.StatusInternalServerError)
			return
		}
		if users == nil {
			users = []models.User{}
		}
		middleware.JSONSuccess(w, users)
	}))

	// WebSocket
	http.HandleFunc("/ws", chatHandler.HandleWebSocket)

	// Health check
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			middleware.JSON(w, http.StatusServiceUnavailable, map[string]interface{}{"status": "unhealthy", "error": "database unreachable"})
			return
		}
		middleware.JSON(w, http.StatusOK, map[string]interface{}{"status": "healthy"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:           ":" + port,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		fmt.Println("\nShutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		}
	}()

	fmt.Printf("🚀 Server starting on http://localhost:%s\n", port)
	fmt.Printf("📡 WebSocket available at ws://localhost:%s/ws\n", port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
	fmt.Println("Server stopped gracefully")
}
