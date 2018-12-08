// Package frontmatter extracts front matter from files and stores it as file metadata.
package frontmatter

import (
	fm "github.com/FooSoft/frontmatter"
	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/extension"
)

// Frontmatter chainable plugin context.
type Frontmatter interface {
	goldsmith.Plugin
	goldsmith.Initializer
	goldsmith.Processor
}

// New creates a new instance of the Frontmatter plugin.
func New() Frontmatter {
	return new(frontmatter)
}

type frontmatter struct{}

func (*frontmatter) Name() string {
	return "frontmatter"
}

func (*frontmatter) Initialize(context *goldsmith.Context) ([]goldsmith.Filter, error) {
	return []goldsmith.Filter{extension.New(".md", ".markdown", ".rst", ".html", ".htm")}, nil
}

func (*frontmatter) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	meta, body, err := fm.Parse(inputFile)
	if err != nil {
		return err
	}

	outputFile := context.CreateFileFromData(inputFile.Path(), body.Bytes())
	outputFile.Meta = inputFile.Meta
	for name, value := range meta {
		outputFile.Meta[name] = value
	}

	context.DispatchFile(outputFile)
	return nil
}
