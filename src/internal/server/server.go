package server

import (
	"net/http"

	"github.com/pkg/errors"
	"go.pkg.littleman.co/library/internal/book"
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
	httpBook, err := book.New(book.WithBook("/usr/share/library/incident-management.epub"))

	if err != nil {
		return errors.Wrap(err, "unable to create http book")
	}

	http.HandleFunc("/healthz", handlers.NoContent)
	http.HandleFunc("/", httpBook.Handler)

	return http.ListenAndServe(s.address, nil)
}
