package middleware

import (
	"log"
	"net/http"

	"github.com/felixge/httpsnoop"
)

// Logging is middleware that logs the HTTP requests & responses
type Logging struct {
}

// NewLogging creates a new logging  middleware
func NewLogging(options ...func(l *Logging) error) (*Logging, error) {
	return &Logging{}, nil
}

// Middleware returns the function that is executed as part of the HTTP middlewares stack
func (l Logging) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		m := httpsnoop.CaptureMetrics(next, w, r)

		log.Printf(
			"%s %s (code=%d dt=%s written=%d)",
			r.Method,
			r.URL,
			m.Code,
			m.Duration,
			m.Written,
		)
	})
}
