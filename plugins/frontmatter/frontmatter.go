// Package frontmatter extracts front matter from files and stores it as file metadata.
package frontmatter

import (
	fm "github.com/FooSoft/frontmatter"
	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/extension"
)

// Frontmatter chainable plugin context.
type FrontMatter struct{}

// New creates a new instance of the Frontmatter plugin.
func New() *FrontMatter {
	return new(FrontMatter)
}

func (*FrontMatter) Name() string {
	return "frontmatter"
}

func (*FrontMatter) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	return extension.New(".md", ".markdown", ".rst", ".html", ".htm"), nil
}

func (*FrontMatter) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
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
