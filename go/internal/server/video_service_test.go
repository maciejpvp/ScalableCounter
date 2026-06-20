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

	// 1. Create a video
	video, err := service.CreateVideo(ctx)
	if err != nil {
		t.Fatalf("failed to create video: %v", err)
	}
	// Extract raw ID from the returned video (which starts with VIDEO#)
	if len(video.VideoID) <= 6 {
		t.Fatalf("unexpected video ID structure: %s", video.VideoID)
	}
	rawID := video.VideoID[6:] // strip "VIDEO#"

	// 2. Perform some rapid likes
	likesCount := 5
	for i := 0; i < likesCount; i++ {
		err := service.LikeVideo(ctx, rawID)
		if err != nil {
			t.Fatalf("failed to like video: %v", err)
		}
	}

	// 3. Immediately read the video - it should contain the buffered likes (5)
	readVideo, err := service.GetVideo(ctx, rawID)
	if err != nil {
		t.Fatalf("failed to get video: %v", err)
	}
	if readVideo.Likes != 5 {
		t.Errorf("expected 5 likes from read-through, got %d", readVideo.Likes)
	}

	// 4. Verify that the DB hasn't been written to yet (likes in DB should still be 0)
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

	// 5. Wait for the flush interval to pass (flushInterval is 1s, wait 1.5s)
	time.Sleep(1500 * time.Millisecond)

	// 6. Verify that the DB now has the flushed likes (5)
	dbVideoAfter, err := repo.GetVideo(ctx, rawID)
	if err != nil {
		t.Fatalf("failed to read direct from repo after flush: %v", err)
	}
	if dbVideoAfter.Likes != 5 {
		t.Errorf("expected 5 likes in DB after flush, got %d", dbVideoAfter.Likes)
	}
}
