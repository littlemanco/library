package problems

import (
	"encoding/hex"
	"fmt"
	"hash/adler32"
	"strings"
)

// FactoryTypeIDTemplate is thestring that is made part of the URI template that will get replaced
const FactoryTypeIDTemplate = "__ID__"

const (
	// AudienceConsumer are the end users of this service.
	AudienceConsumer = iota

	// AudienceAPIUser are users who are accessing this service via an API
	AudienceAPIUser = iota

	// AudienceDeveloper are users who are maintaining this software
	AudienceDeveloper = iota
)

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

	// The users who are the intended target of this problem message
	Audience []int
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

// IsFor Allows Checking for whom this error is intended
func (p Problem) IsFor(Audience int) bool {
	for _, x := range p.Audience {
		if Audience == x {
			return true
		}
	}

	return false
}

// New Creates a new problem object
func New(Type string, Title, Message string, Audience []int) *Problem {
	return &Problem{
		Type:        Type,
		Title:       Title,
		Description: Message,
		Audience:    Audience,
	}
}

// WithEverything creates a complete problem object
func (f Factory) WithEverything(Title string, Description string, Audience []int) *Problem {
	h := adler32.New()
	h.Write([]byte(Title))

	URI := strings.Replace(f.URITemplate, FactoryTypeIDTemplate, hex.EncodeToString(h.Sum(nil)), -1)

	return New(
		URI,
		Title,
		Description,
		Audience,
	)
}

// WithTitle creates a new problem with only a title
func (f Factory) WithTitle(Title string) *Problem {
	return f.WithEverything(
		Title,
		"",
		[]int{AudienceDeveloper},
	)
}

// WithTitleAudience is for denoting when this problem is for a specific audience (usually non-developer)
func (f Factory) WithTitleAudience(Title string, Audience []int) *Problem {
	return f.WithEverything(
		Title,
		"",
		Audience,
	)
}
