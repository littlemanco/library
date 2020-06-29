package book

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

type renderer func(io.Reader, io.Writer) error

// HTTPBook is a book that will be returned over HTTP.
type HTTPBook struct {
	Book
}

const (
	extTypeXHTML = ".xhtml"
)

// Handler is the HTTP handler that serves the appropriate book content
func (h Book) Handler(w http.ResponseWriter, r *http.Request) {
	path := r.RequestURI

	// In the case this is the root, transform the root into the nav file.
	if path == "/" {
		path = "/nav.xhtml"
	}

	// Check if the file is in the book
	exists := false
	for _, f := range h.EPub.Files() {
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
	file, err := h.EPub.Open(path)
	ext := filepath.Ext(path)

	if err != nil {
		renderError(err, w)
		return
	}

	w.Header().Set("Content-Type", mime.TypeByExtension(ext))

	switch ext {
	case extTypeXHTML:
		if err := renderHTML(file, w); err != nil {
			renderError(err, w)
		}
	default:
		if err := renderSimple(file, w); err != nil {
			renderError(err, w)
		}
	}
}

func renderSimple(h io.Reader, w http.ResponseWriter) error {
	io.Copy(w, h)

	return nil
}

func renderHTML(h io.Reader, w http.ResponseWriter) error {
	xhtmlMobileFriendly := []*html.Node{
		{Type: html.ElementNode, Data: "meta", Attr: []html.Attribute{
			{Key: "name", Val: "viewport"},
			{Key: "content", Val: "width=device-width, initial-scale=1.0"},
		}},
		{Type: html.ElementNode, Data: "style", Attr: []html.Attribute{
			{Key: "type", Val: "text/css"},
		}, FirstChild: &html.Node{
			Type: html.TextNode, Data: `
body {
	display: block;
	margin: 0 auto;
	max-width: 1200px;
	padding: 0 15px !important;
}
`,
		}},

		// Google Analycs
		{Type: html.ElementNode, Data: "script", Attr: []html.Attribute{
			{Key: "async"},
			{Key: "src", Val: "https://www.googletagmanager.com/gtag/js?id=UA-53227254-5"},
		}},
		{Type: html.ElementNode, Data: "script", FirstChild: &html.Node{
			Type: html.TextNode, Data: `
			window.dataLayer = window.dataLayer || [];
			function gtag(){dataLayer.push(arguments);}
			gtag('js', new Date());

			gtag('config', 'UA-53227254-5');
`,
		}},
	}

	// Function to traverse the HTML tree
	// Todo: This should be pulled out, and the pages modified before storing in memory for later access.
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "head" {
			for _, x := range xhtmlMobileFriendly {
				n.AppendChild(x)
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	doc, err := html.Parse(h)

	if err != nil {
		return errors.Wrap(err, "unable to parse html")
	}

	// Modify DOc
	f(doc)
	html.Render(w, doc)

	return nil
}

func renderError(e error, w http.ResponseWriter) {
	// Todo: This should be a better error handler, including logging errors
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(e.Error()))
}
