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
)

func main() {
	// Load configuration
	cfg := config.Load()

	log.Printf("🚀 Starting 図面くん Go Backend on port %s (env: %s)", cfg.Port, cfg.Env)
	log.Printf("🤖 Vision AI: Gemini 2.0 Flash (direct PDF mode)")

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
			"ai":        "gemini-2.0-flash",
		}
		c.JSON(http.StatusOK, status)
	})

	// Serve uploaded PDFs (開発環境用インメモリストア)
	router.GET("/uploads/*filepath", handlers.ServeUploadedFile)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// ── 認証（認証不要） ──────────────────────────────
		authHandler := handlers.NewAuthHandler(supabase)
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/register", authHandler.Register)
			authGroup.GET("/me", middleware.AuthRequired(), authHandler.Me)
		}

		// ── PDF アップロード（認証不要・開発中） ───────────
		pdfHandler := handlers.NewPDFHandler(supabase, cfg)
		v1.POST("/properties/parse", pdfHandler.UploadAndParse)

		// ── 物件 CRUD ──────────────────────────────────────
		propHandler := handlers.NewPropertyHandler(supabase, cfg)
		props := v1.Group("/properties")
		{
			props.GET("", propHandler.ListProperties)
			props.POST("", propHandler.CreateProperty)
			props.DELETE("", propHandler.DeleteAllProperties)
			props.DELETE("/bulk", propHandler.BulkDeleteProperties)
			props.GET("/:id", propHandler.GetProperty)
			props.PATCH("/:id", propHandler.UpdateProperty)
			props.DELETE("/:id", propHandler.DeleteProperty)

			props.POST("/:id/parse", propHandler.TriggerParse)
			props.POST("/:id/call", propHandler.TriggerCall)
			props.POST("/:id/cancel-call", propHandler.CancelCall)
			props.GET("/:id/pdf", pdfHandler.ServePDF)

			callHandler := handlers.NewCallHandler(supabase)
			props.GET("/:id/calls", callHandler.GetCallsByProperty)
		}
	}

	// Start server with graceful shutdown
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: router,
	}

	go func() {
		log.Printf("✅ 図面くん API server running at http://localhost:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server error: %v", err)
		}
	}()

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
