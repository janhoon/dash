package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/janhoon/dash/backend/internal/db"
	"github.com/janhoon/dash/backend/internal/handlers"
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

	// Setup router
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /api/health", handlers.HealthCheck)

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

	log.Println("Server stopped")
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
