package problems

import (
	"encoding/hex"
	"fmt"
	"hash/adler32"
	"strings"
)

// FactoryTypeIDTemplate is thestring that is made part of the URI template that will get replaced
const FactoryTypeIDTemplate = "__ID__"

// Problem is an error that has additional metadata associated with it designed to help users understand and resolve
// it.
//
// Inspired by https://tools.ietf.org/html/rfc7807
type Problem struct {
	// A URI assigned to this problem
	//
	// Accessing this URI *should* describe additional details about this problem
	Type string

	// A short summary of this problem
	//
	// MUST Be unique across this type of problem
	Title string

	// An optional description of the problem
	Description string
}

// Factory allows easy, opinionated problem creation
type Factory struct {
	// A template that will be replaced with a URL
	URITemplate string
}

// Error implements the error interface, allowing this type to be returned as an error.
func (p Problem) Error() string {
	return p.String()
}

// String allows this interface to be rendered as a string
func (p Problem) String() string {
	if len(p.Description) == 0 {
		return fmt.Sprintf("%s (%s)", p.Title, p.Type)
	}

	return fmt.Sprintf("%s: %s (%s)", p.Title, p.Description, p.Type)
}

// New Creates a new problem object
func New(Type string, Title, Message string) *Problem {
	return &Problem{
		Type:        Type,
		Title:       Title,
		Description: Message,
	}
}

// WithEverything creates a complete problem object
func (f Factory) WithEverything(Title string, Description string) *Problem {
	h := adler32.New()
	h.Write([]byte(Title))

	URI := strings.Replace(f.URITemplate, FactoryTypeIDTemplate, hex.EncodeToString(h.Sum(nil)), -1)

	return New(
		URI,
		Title,
		Description,
	)
}

// WithTitle creates a new problem with only a title
func (f Factory) WithTitle(Title string) *Problem {
	return f.WithEverything(
		Title,
		"",
	)
}
