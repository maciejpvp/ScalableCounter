package server

import (
	"github.com/go-chi/chi/v5"
)

func (s *Server) videoRoutes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", s.videoController.CreateVideoHandler)
	r.Get("/{id}", s.videoController.GetVideoHandler)
	r.Post("/{id}/like", s.videoController.LikeVideoHandler)

	return r
}
