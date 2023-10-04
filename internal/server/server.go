package server

import (
	"net/http"

	"github.com/Dmitrevicz/gometrics/internal/storage/memstorage"
)

type server struct {
	router   http.Handler
	handlers *Handlers
	// storage  *Storage
}

func New() *server {
	storage := memstorage.New()
	s := server{
		router:   http.NewServeMux(),
		handlers: NewHandlers(storage),
		// storage:  storage,
	}

	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, s.handlers.Update)
	mux.HandleFunc(`/`, s.handlers.GetAllMetrics)

	s.router = mux

	return &s
}

// requires router of type *http.ServeMux
// func (s *server) configureRouter() {
// 	s.Router.HandleFunc(`/update/`, internal.UpdateHandler)
// }

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
