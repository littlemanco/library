package book

import (
	"github.com/kapmahc/epub"
	"github.com/pkg/errors"
)

// Book is a primitive that describes an ePub
type Book struct {
	// EPub is the actual book being served
	EPub *epub.Book
}

// New creates a new Book entity
func New(options ...func(*Book) error) (*Book, error) {
	b := &Book{}

	for _, o := range options {
		if err := o(b); err != nil {
			return nil, errors.Wrap(err, "unable to persist option to Book")
		}
	}

	// Check required properties
	if b.EPub == nil {
		return nil, errors.New("cannot create http book: no book supplied")
	}

	return b, nil
}

// WithEPUB adds the book to the Book
func WithEPUB(path string) func(*Book) error {
	return func(h *Book) error {
		book, err := epub.Open(path)

		if err != nil {
			return errors.Wrap(err, "unable to open book")
		}

		h.EPub = book

		return nil
	}
}
