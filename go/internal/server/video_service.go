package server

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"scalable_counter/internal/database"
)

type VideoService struct {
	repo          *database.VideoRepository
	likesBuffer   map[string]int
	bufferMu      sync.RWMutex
	flushInterval time.Duration
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

func NewVideoService(repo *database.VideoRepository) *VideoService {
	s := &VideoService{
		repo:          repo,
		likesBuffer:   make(map[string]int),
		flushInterval: 1 * time.Second,
		stopChan:      make(chan struct{}),
	}
	s.wg.Add(1)
	go s.startFlushWorker()
	return s
}

func (s *VideoService) CreateVideo(ctx context.Context) (*database.VideoRecord, error) {
	videoID := fmt.Sprintf("%d", time.Now().UnixNano())

	record, err := s.repo.PutVideo(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to create video in repository: %w", err)
	}

	return record, nil
}

func (s *VideoService) GetVideo(ctx context.Context, videoID string) (*database.VideoRecord, error) {
	record, err := s.repo.GetVideo(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video in repository: %w", err)
	}

	if record != nil {
		s.bufferMu.RLock()
		bufferedLikes := s.likesBuffer[videoID]
		s.bufferMu.RUnlock()

		record.Likes += bufferedLikes
	}

	return record, nil
}

func (s *VideoService) LikeVideo(ctx context.Context, videoID string) error {
	s.bufferMu.Lock()
	s.likesBuffer[videoID]++
	s.bufferMu.Unlock()
	return nil
}

func (s *VideoService) startFlushWorker() {
	defer s.wg.Done()
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.flush()
		case <-s.stopChan:
			s.flush()
			return
		}
	}
}

func (s *VideoService) flush() {
	s.bufferMu.Lock()
	if len(s.likesBuffer) == 0 {
		s.bufferMu.Unlock()
		return
	}
	pendingLikes := s.likesBuffer
	s.likesBuffer = make(map[string]int)
	s.bufferMu.Unlock()

	for videoID, likes := range pendingLikes {
		ctx := context.Background()
		err := s.repo.IncrementVideoLikes(ctx, videoID, likes)
		if err != nil {
			log.Printf("failed to flush %d likes for video %s: %v", likes, videoID, err)
			s.bufferMu.Lock()
			s.likesBuffer[videoID] += likes
			s.bufferMu.Unlock()
		}
	}
}

func (s *VideoService) Close() {
	close(s.stopChan)
	s.wg.Wait()
}