package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type VideoController struct {
	service *VideoService
}

func NewVideoController(service *VideoService) *VideoController {
	return &VideoController{service: service}
}

func (c *VideoController) CreateVideoHandler(w http.ResponseWriter, r *http.Request) {
	record, err := c.service.CreateVideo(r.Context())
	if err != nil {
		log.Printf("error creating video: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	
	jsonResp, _ := json.Marshal(record)
	_, _ = w.Write(jsonResp)
}

func (c *VideoController) GetVideoHandler(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "id")
	record, err := c.service.GetVideo(r.Context(), videoID)
	if err != nil {
		log.Printf("error getting video: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if record == nil {
		http.Error(w, "video not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	jsonResp, _ := json.Marshal(record)
	_, _ = w.Write(jsonResp)
}

func (c *VideoController) LikeVideoHandler(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "id")
	err := c.service.LikeVideo(r.Context(), videoID)
	if err != nil {
		log.Printf("error liking video: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	jsonResp, _ := json.Marshal(map[string]string{"message": "Video liked successfully"})
	_, _ = w.Write(jsonResp)
}