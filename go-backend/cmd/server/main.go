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

	"github.com/gin-gonic/gin"
	"github.com/zumen-kun/go-backend/internal/api/handlers"
	"github.com/zumen-kun/go-backend/internal/api/middleware"
	"github.com/zumen-kun/go-backend/internal/config"
	"github.com/zumen-kun/go-backend/internal/db"
	redisClient "github.com/zumen-kun/go-backend/internal/redis"
)

func main() {
	// Load configuration
	cfg := config.Load()

	log.Printf("🚀 Starting 図面くん Go Backend on port %s (env: %s)", cfg.Port, cfg.Env)

	// Set Gin mode
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Supabase client
	supabase, err := db.NewSupabaseClient(cfg)
	if err != nil {
		log.Printf("⚠️  Supabase client warning: %v (continuing without DB)", err)
		supabase = &db.SupabaseClient{}
	}

	// Initialize Redis client
	redis, err := redisClient.NewClient(cfg)
	if err != nil {
		log.Printf("⚠️  Redis client warning: %v (continuing without Redis)", err)
		redis = nil
	}

	// Initialize task producer
	var producer *redisClient.TaskProducer
	if redis != nil {
		producer = redisClient.NewTaskProducer(redis)
		defer redis.Close()
	}

	// Setup Gin router
	router := gin.Default()

	// Apply middlewares
	router.Use(middleware.CORS())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		status := gin.H{
			"status":    "ok",
			"service":   "zumen-kun-go-backend",
			"timestamp": time.Now().UTC(),
		}

		if redis != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := redis.GetClient().Ping(ctx).Err(); err != nil {
				status["redis"] = "disconnected"
			} else {
				status["redis"] = "connected"
			}
		} else {
			status["redis"] = "not configured"
		}

		c.JSON(http.StatusOK, status)
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// PDF upload & parse
		pdfHandler := handlers.NewPDFHandler(supabase, producer, cfg)
		v1.POST("/properties/parse", pdfHandler.UploadAndParse)

		// Property CRUD
		propHandler := handlers.NewPropertyHandler(supabase, producer)
		props := v1.Group("/properties")
		{
			props.GET("", propHandler.ListProperties)
			props.POST("", propHandler.CreateProperty)
			props.GET("/:id", propHandler.GetProperty)
			props.PATCH("/:id", propHandler.UpdateProperty)
			props.DELETE("/:id", propHandler.DeleteProperty)

			// Action endpoints
			props.POST("/:id/parse", propHandler.TriggerParse)
			props.POST("/:id/call", propHandler.TriggerCall)

			// Call records
			callHandler := handlers.NewCallHandler(supabase)
			props.GET("/:id/calls", callHandler.GetCallsByProperty)
		}
	}

	// Start server with graceful shutdown
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: router,
	}

	// Start in goroutine
	go func() {
		log.Printf("✅ 図面くん API server running at http://localhost:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⏳ Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited")
}
