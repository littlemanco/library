package book

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/kapmahc/epub"
	"github.com/pkg/errors"
)

// HTTPBook represents a "servable book"
type HTTPBook struct {

	// Book is the actual book being served
	Book *epub.Book
}

// Handler is the HTTP handler that serves the appropriate book content
func (h HTTPBook) Handler(w http.ResponseWriter, r *http.Request) {
	path := r.RequestURI

	// In the case this is the root, transform the root into the nav file.
	if path == "/" {
		path = "/nav.xhtml"
	}

	// Check if the file is in the book
	exists := false
	for _, f := range h.Book.Files() {
		if fmt.Sprintf("EPUB%s", path) == f {
			exists = true
		}
	}

	// If the file is not th ere, return 404
	if exists == false {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Open the file for reading
	file, err := h.Book.Open(path)

	if err != nil {
		// Todo: Logging Here
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(file)
	w.Write(buf.Bytes())
}

// New creates a new HTTPBook entity
func New(options ...func(*HTTPBook) error) (*HTTPBook, error) {
	b := &HTTPBook{}

	for _, o := range options {
		if err := o(b); err != nil {
			return nil, errors.Wrap(err, "unable to persist option to Book")
		}
	}

	// Check required properties
	if b.Book == nil {
		return nil, errors.New("cannot create http book: no book supplied")
	}

	return b, nil
}

// WithBook adds the book to the HTTPBook
func WithBook(path string) func(*HTTPBook) error {
	return func(h *HTTPBook) error {
		book, err := epub.Open(path)

		if err != nil {
			return errors.Wrap(err, "unable to open book")
		}

		h.Book = book

		return nil
	}
}
