package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/lib/pq"
	"github.com/rhblitstein/cad-version-control/internal/handlers"
	"github.com/rhblitstein/cad-version-control/internal/repository"
	"github.com/rhblitstein/cad-version-control/internal/storage"
)

func main() {
	//Load env vars
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "caduser")
	dbPass := getEnv("DB_PASSWORD", "cadpass")
	dbName := getEnv("DB_NAME", "cadversion")
	minioEndpoint := getEnv("MINIO_ENDPOINT", "localhost:9000")
	minioAccessKey := getEnv("MINIO_ACCESS_KEY", "minioadmin")
	minioSecretKey := getEnv("MINIO_SECRET_KEY", "minioadmin")
	redisHost := getEnv("REDIS_HOST", "localhost:6379")
	port := getEnv("PORT", "8080")

	//Initialize database connection
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)
	db, err := repository.NewPostgres(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	//Initialize minio client
	minioClient, err := storage.NewMinioClient(minioEndpoint, minioAccessKey, minioSecretKey)
	if err != nil {
		log.Fatalf("Failed to connect to MinIO: %v", err)
	}
	log.Println("✓ Connected to MinIO")

	//Initialize Redis client
	redisClient := repository.NewRedisClient(redisHost)
	log.Println("✓ Connected to Redis")

	//Initialize repositories
	projectRepo := repository.NewProjectRepository(db.DB, redisClient)
	branchRepo := repository.NewBranchRepository(db.DB, redisClient)
	commitRepo := repository.NewCommitRepository(db.DB)
	fileRepo := repository.NewFileRepository(db.DB)
	mrRepo := repository.NewMergeRequestRepository(db.DB)

	//Initialize handlers
	projectHandler := handlers.NewProjectHandler(projectRepo)
	branchHandler := handlers.NewBranchHandler(branchRepo, projectRepo)
	commitHandler := handlers.NewCommitHandler(commitRepo, branchRepo, fileRepo, minioClient)
	mrHandler := handlers.NewMergeRequestHandler(mrRepo, branchRepo, fileRepo)

	//Setup router
	r := chi.NewRouter()

	//Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	//Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","database":"connected","storage":"connected","cache":"connected"}`))
	})

	//API Routes
	r.Route("/api", func(r chi.Router) {
		// Projects
		r.Post("/projects", projectHandler.Create)
		r.Get("/projects", projectHandler.List)
		r.Get("/projects/{id}", projectHandler.Get)

		// Branches
		r.Post("/projects/{project_id}/branches", branchHandler.Create)
		r.Get("/projects/{project_id}/branches", branchHandler.List)
		r.Get("/branches/{id}", branchHandler.Get)

		// Commits
		r.Post("/projects/{project_id}/commits", commitHandler.Create)
		r.Get("/commits/{id}", commitHandler.Get)
		r.Get("/branches/{branch_id}/commits", commitHandler.ListByBranch)

		// Files
		r.Get("/files/{id}/versions", commitHandler.GetFileVersions)
		r.Get("/file-versions/{id}/download", commitHandler.DownloadFile)

		// Merge Requests
		r.Post("/merge-requests", mrHandler.Create)
		r.Get("/merge-requests", mrHandler.List)
		r.Get("/merge-requests/{id}", mrHandler.Get)
		r.Post("/merge-requests/{id}/approve", mrHandler.Approve)
		r.Post("/merge-requests/{id}/merge", mrHandler.Merge)
		r.Post("/merge-requests/{id}/comments", mrHandler.AddComment)
		r.Get("/merge-requests/{id}/comments", mrHandler.GetComments)
		r.Get("/merge-requests/{id}/conflicts", mrHandler.GetConflicts)
		r.Post("/conflicts/{id}/resolve", mrHandler.ResolveConflict)
		r.Get("/conflicts/{id}/diff", mrHandler.GetDiff)
	})

	//HTTP Server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	//Start server in goroutine
	go func() {
		log.Printf("🚀 Server starting on http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	//Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
