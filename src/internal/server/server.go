package server

import (
	"net/http"

	"go.pkg.littleman.co/library/internal/server/handlers"
)

// Server is the entity that listens to HTTP requests and responds
type Server struct {
	address string
}

// New returns a new server instance
func New(options ...func(*Server) error) (*Server, error) {
	return &Server{
		address: "0.0.0.0:8080",
	}, nil
}

// Serve starts the server
func (s Server) Serve() error {
	http.HandleFunc("/healthz", handlers.NoContent)

	return http.ListenAndServe(s.address, nil)
}
