package handlers

import "net/http"

// NoContent returns a simple response indicating that there is nothing here, but that's fine.
func NoContent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
