package main

import (
	"context"
	"fmt"
	"log"

	"scalable_counter/internal/database"
)

func main() {
	db, err := database.NewDB(context.Background())
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}
	repo := database.NewVideoRepository(db)
	res, err := repo.PutVideo(context.Background(), "test-video-123")
	if err != nil {
		log.Fatalf("PutVideo error: %v", err)
	}
	fmt.Printf("PutVideo success: %+v\n", res)
}
