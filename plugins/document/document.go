// Package document enables simple HTML modification via callback via
// "goquery", an API similar to "jquery". This plugin is particularly useful
// adding classes to elements and performing other cleanup tasks which are too
// case-specific to warrant the creation of a new plugin.
package document

import (
	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/wildcard"
	"github.com/PuerkitoBio/goquery"
)

// Processor callback function to modify documents.
type Processor func(*goquery.Document) error

// Document plugin context.
type Document struct {
	callback Processor
}

// New creates a new instance of the Document plugin.
func New(callback Processor) *Document {
	return &Document{callback}
}

func (*Document) Name() string {
	return "document"
}

func (*Document) Initialize(context *goldsmith.Context) error {
	context.Filter(wildcard.New("**/*.html", "**/*.htm"))
	return nil
}

func (plugin *Document) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	doc, err := goquery.NewDocumentFromReader(inputFile)
	if err != nil {
		return err
	}

	if err := plugin.callback(doc); err != nil {
		return err
	}

	html, err := doc.Html()
	if err != nil {
		return err
	}

	outputFile := context.CreateFileFromData(inputFile.Path(), []byte(html))
	outputFile.Meta = inputFile.Meta
	context.DispatchFile(outputFile)
	return nil
}
