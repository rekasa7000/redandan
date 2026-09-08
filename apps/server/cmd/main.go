package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"reliva/server/internal/config"
	"reliva/server/internal/db"
	"reliva/server/internal/handlers"
	"reliva/server/internal/routes"
)

func main() {
	// Load .env in development; ignore error in production (env vars are injected).
	_ = godotenv.Load()

	cfg := config.Load()

	database, disconnect, err := db.Connect(cfg.MongoURI, cfg.DBName)
	if err != nil {
		log.Fatalf("mongodb: failed to connect: %v", err)
	}
	defer disconnect()

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.Default()

	h := handlers.New(database, cfg)
	routes.Register(r, h, cfg)

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine so the graceful shutdown logic below can run.
	go func() {
		log.Printf("server listening on :%s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	// Block until SIGINT or SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced shutdown: %v", err)
	}

	log.Println("server stopped")
}
