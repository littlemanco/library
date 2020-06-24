package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.pkg.littleman.co/library/internal/book"
	"go.pkg.littleman.co/library/internal/server/handlers"
)

// Server is the entity that listens to HTTP requests and responds
type Server struct {
	address  string
	bookPath string
}

// New returns a new server instance
func New(options ...func(*Server) error) (*Server, error) {
	s := &Server{
		address:  "0.0.0.0:8080",
		bookPath: "/book.epub",
	}

	for _, o := range options {
		if err := o(s); err != nil {
			return nil, errors.Wrap(err, "unable to set option on configuration")
		}
	}

	return s, nil
}

// WithBook allows supplying the book path to the server
func WithBook(path string) func(*Server) error {
	return func(s *Server) error {
		s.bookPath = path

		return nil
	}
}

// Serve starts the server
func (s Server) Serve() error {
	httpBook, err := book.New(book.WithBook(s.bookPath))

	if err != nil {
		return errors.Wrap(err, "unable to create http book")
	}

	r := mux.NewRouter()

	// Bind the routes
	r.HandleFunc("/healthz", handlers.NoContent)
	r.HandleFunc("/", httpBook.Handler)

	// Set the router
	http.Handle("/", r)

	return http.ListenAndServe(s.address, nil)
}
