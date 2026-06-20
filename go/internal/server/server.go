package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"scalable_counter/internal/database"

	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port            int
	db              *database.DB
	videoController *VideoController
	videoService    *VideoService
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	// Initialize database
	db, err := database.NewDB(context.Background())
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	// Initialize Video related dependencies
	videoRepo := database.NewVideoRepository(db)
	videoService := NewVideoService(videoRepo)
	videoController := NewVideoController(videoService)

	NewServer := &Server{
		port:            port,
		db:              db,
		videoController: videoController,
		videoService:    videoService,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	server.RegisterOnShutdown(func() {
		log.Println("Shutting down video service (flushing write-behind cache)...")
		videoService.Close()
		log.Println("Video service shut down successfully.")
	})

	return server
}
