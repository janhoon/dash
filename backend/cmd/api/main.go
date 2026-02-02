package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/janhoon/dash/backend/internal/auth"
	"github.com/janhoon/dash/backend/internal/db"
	"github.com/janhoon/dash/backend/internal/handlers"
	"github.com/janhoon/dash/backend/internal/valkey"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://dash:dash@localhost:5432/dash?sslmode=disable"
	}

	// Get Prometheus URL from environment
	prometheusURL := os.Getenv("PROMETHEUS_URL")
	if prometheusURL == "" {
		prometheusURL = "http://localhost:9090"
	}

	// Connect to database
	pool, err := db.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Run migrations
	if err := db.RunMigrations(context.Background(), pool); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize JWT manager
	jwtManager, err := auth.NewJWTManager()
	if err != nil {
		log.Fatalf("Failed to initialize JWT manager: %v", err)
	}

	// Initialize Valkey client (optional - refresh tokens won't work without it)
	valkeyClient, err := valkey.NewClient()
	if err != nil {
		log.Printf("Warning: Valkey not available, refresh tokens disabled: %v", err)
	} else {
		defer valkeyClient.Close()
		log.Println("Valkey connected successfully")
	}

	// Setup router
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /api/health", handlers.HealthCheck)

	// Auth routes
	var rdb *redis.Client
	if valkeyClient != nil {
		rdb = valkeyClient.GetRedis()
	}
	authHandler := handlers.NewAuthHandler(pool, jwtManager, rdb)
	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("GET /api/auth/me", auth.RequireAuth(jwtManager, authHandler.Me))
	mux.HandleFunc("GET /api/auth/me/methods", auth.RequireAuth(jwtManager, authHandler.GetAuthMethods))
	mux.HandleFunc("DELETE /api/auth/me/methods/{id}", auth.RequireAuth(jwtManager, authHandler.UnlinkAuthMethod))
	mux.HandleFunc("POST /api/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /api/auth/logout", authHandler.Logout)
	mux.HandleFunc("POST /api/auth/logout-all", auth.RequireAuth(jwtManager, authHandler.LogoutAll))

	// Google SSO routes
	googleSSOHandler := handlers.NewGoogleSSOHandler(pool, jwtManager)
	mux.HandleFunc("GET /api/auth/google/login", googleSSOHandler.Login)
	mux.HandleFunc("GET /api/auth/google/callback", googleSSOHandler.Callback)
	mux.HandleFunc("POST /api/orgs/{id}/sso/google", auth.RequireAuth(jwtManager, googleSSOHandler.ConfigureSSO))
	mux.HandleFunc("GET /api/orgs/{id}/sso/google", auth.RequireAuth(jwtManager, googleSSOHandler.GetSSOConfig))

	// Microsoft SSO routes
	microsoftSSOHandler := handlers.NewMicrosoftSSOHandler(pool, jwtManager)
	mux.HandleFunc("GET /api/auth/microsoft/login", microsoftSSOHandler.Login)
	mux.HandleFunc("GET /api/auth/microsoft/callback", microsoftSSOHandler.Callback)
	mux.HandleFunc("POST /api/orgs/{id}/sso/microsoft", auth.RequireAuth(jwtManager, microsoftSSOHandler.ConfigureSSO))
	mux.HandleFunc("GET /api/orgs/{id}/sso/microsoft", auth.RequireAuth(jwtManager, microsoftSSOHandler.GetSSOConfig))

	// Dashboard routes
	dashboardHandler := handlers.NewDashboardHandler(pool)
	mux.HandleFunc("POST /api/dashboards", dashboardHandler.Create)
	mux.HandleFunc("GET /api/dashboards", dashboardHandler.List)
	mux.HandleFunc("GET /api/dashboards/{id}", dashboardHandler.Get)
	mux.HandleFunc("PUT /api/dashboards/{id}", dashboardHandler.Update)
	mux.HandleFunc("DELETE /api/dashboards/{id}", dashboardHandler.Delete)

	// Panel routes
	panelHandler := handlers.NewPanelHandler(pool)
	mux.HandleFunc("POST /api/dashboards/{id}/panels", panelHandler.Create)
	mux.HandleFunc("GET /api/dashboards/{id}/panels", panelHandler.ListByDashboard)
	mux.HandleFunc("PUT /api/panels/{id}", panelHandler.Update)
	mux.HandleFunc("DELETE /api/panels/{id}", panelHandler.Delete)

	// Prometheus data source routes
	prometheusHandler := handlers.NewPrometheusHandler(prometheusURL)
	mux.HandleFunc("GET /api/datasources/prometheus/query", prometheusHandler.Query)
	mux.HandleFunc("GET /api/datasources/prometheus/metrics", prometheusHandler.Metrics)
	mux.HandleFunc("GET /api/datasources/prometheus/labels", prometheusHandler.Labels)
	mux.HandleFunc("GET /api/datasources/prometheus/label/{name}/values", prometheusHandler.LabelValues)

	// Apply CORS middleware
	handler := corsMiddleware(mux)

	// Create server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Println("Starting server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
