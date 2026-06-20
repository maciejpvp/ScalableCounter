package server

import (
	"context"
	"os"
	"testing"
	"time"

	"scalable_counter/internal/database"
)

func TestVideoServiceWriteBehind(t *testing.T) {
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if endpoint == "" {
		t.Skip("Skipping integration test; DYNAMODB_ENDPOINT not set")
	}

	ctx := context.Background()
	db, err := database.NewDB(ctx)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	repo := database.NewVideoRepository(db)
	service := NewVideoService(repo)
	defer service.Close()

	video, err := service.CreateVideo(ctx)
	if err != nil {
		t.Fatalf("failed to create video: %v", err)
	}

	if len(video.VideoID) <= 6 {
		t.Fatalf("unexpected video ID structure: %s", video.VideoID)
	}
	rawID := video.VideoID[6:] // strip "VIDEO#"

	likesCount := 5
	for i := 0; i < likesCount; i++ {
		err := service.LikeVideo(ctx, rawID)
		if err != nil {
			t.Fatalf("failed to like video: %v", err)
		}
	}

	readVideo, err := service.GetVideo(ctx, rawID)
	if err != nil {
		t.Fatalf("failed to get video: %v", err)
	}
	if readVideo.Likes != 5 {
		t.Errorf("expected 5 likes from read-through, got %d", readVideo.Likes)
	}

	dbVideo, err := repo.GetVideo(ctx, rawID)
	if err != nil {
		t.Fatalf("failed to read direct from repo: %v", err)
	}
	if dbVideo == nil {
		t.Fatalf("expected video to exist in DB")
	}
	if dbVideo.Likes != 0 {
		t.Errorf("expected 0 likes directly in DB before flush, got %d", dbVideo.Likes)
	}

	time.Sleep(1500 * time.Millisecond)

	dbVideoAfter, err := repo.GetVideo(ctx, rawID)
	if err != nil {
		t.Fatalf("failed to read direct from repo after flush: %v", err)
	}
	if dbVideoAfter.Likes != 5 {
		t.Errorf("expected 5 likes in DB after flush, got %d", dbVideoAfter.Likes)
	}
}
