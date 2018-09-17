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
	return new(frontmatterPlugin)
}

type frontmatterPlugin struct {
}

func (*frontmatterPlugin) Name() string {
	return "frontmatter"
}

func (*frontmatterPlugin) Initialize(context goldsmith.Context) ([]goldsmith.Filter, error) {
	return []goldsmith.Filter{extension.New(".md", ".markdown", ".rst", ".html", ".htm")}, nil
}

func (*frontmatterPlugin) Process(context goldsmith.Context, f goldsmith.File) error {
	meta, body, err := fm.Parse(f)
	if err != nil {
		return err
	}

	nf := goldsmith.NewFileFromData(f.Path(), body.Bytes(), f.ModTime())
	nf.InheritValues(f)
	for name, value := range meta {
		nf.SetValue(name, value)
	}
	context.DispatchFile(nf)

	return nil
}
