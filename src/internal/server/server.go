package server

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.pkg.littleman.co/library/internal/book"
	"go.pkg.littleman.co/library/internal/server/handlers"
	"go.pkg.littleman.co/library/internal/server/middleware"
)

// OIDCConfig is the authentication configuration for an OIDC Server
type OIDCConfig struct {
	Provider     string
	ClientID     string
	ClientSecret string
	RedirectURL  *url.URL
	ClaimSets    []middleware.OIDCClaimSet
}

// Server is the entity that listens to HTTP requests and responds
type Server struct {
	address  string
	bookPath string

	middleware []mux.MiddlewareFunc
}

// Option is a function that modifies servers behaviour
type Option func(*Server) error

// New returns a new server instance
func New(options ...Option) (*Server, error) {
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

// WithOIDCAuthentication modifies the library to authenticate users against an OIDC Endpoint
func WithOIDCAuthentication(config *OIDCConfig) func(*Server) error {
	return func(s *Server) error {
		options := []middleware.OIDCAuthConfiguration{}

		for _, c := range config.ClaimSets {
			options = append(options, middleware.WithClaimSet(c))
		}

		auth, err := middleware.NewOidcAuth(
			config.Provider,
			config.ClientID,
			config.ClientSecret,
			config.RedirectURL,
			options...,
		)

		if err != nil {
			return errors.Wrap(err, "unable to create OIDC Middleware")
		}

		s.middleware = append(s.middleware, auth.Middleware)

		return nil
	}
}

// WithLogging enables the logging on the server component
func WithLogging() func(*Server) error {
	return func(s *Server) error {
		m, e := middleware.NewLogging()

		if e != nil {
			return errors.Wrap(e, "unable to set up logging middleware")
		}

		s.middleware = append(s.middleware, m.Middleware)

		return nil
	}
}

// Serve starts the server
func (s Server) Serve() error {
	Book, err := book.New(book.WithEPUB(s.bookPath))

	if err != nil {
		return errors.Wrap(err, "unable to create http book")
	}

	// Specialized routes
	http.HandleFunc("/healthz", handlers.NoContent)

	// Normal Routes
	r := mux.NewRouter()

	// Bind the routes)
	r.Use(s.middleware...)
	r.PathPrefix("/").HandlerFunc(Book.Handler)

	// Set router to HTTP server
	http.Handle("/", r)

	return http.ListenAndServe(s.address, nil)
}
