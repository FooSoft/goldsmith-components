// Package document allows for modification of HTML document structure.
package document

import (
	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/extension"
	"github.com/PuerkitoBio/goquery"
)

// Document plugin context.
type Document interface {
	goldsmith.Plugin
	goldsmith.Initializer
	goldsmith.Processor
}

// Processor callback function to modify documents.
type Processor func(*goquery.Document) error

// New creates a new instance of the Dom plugin.
func New(callback Processor) Document {
	return &document{callback}
}

type document struct {
	callback Processor
}

func (*document) Name() string {
	return "document"
}

func (*document) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	return extension.New(".html", ".htm"), nil
}

func (document *document) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	doc, err := goquery.NewDocumentFromReader(inputFile)
	if err != nil {
		return err
	}

	if err := document.callback(doc); err != nil {
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
