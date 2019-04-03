// Package document enables simple HTML modification via callback.
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

// New creates a new instance of the Dom plugin.
// The provided callback will be invoked for all HTML documents.
func New(callback Processor) *Document {
	return &Document{callback}
}

func (*Document) Name() string {
	return "document"
}

func (*Document) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	return wildcard.New("**/*.html", "**/*.htm"), nil
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
