// Package minify removes superfluous data from a variety of web formats.
package minify

import (
	"bytes"
	"path/filepath"

	min "github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/svg"
	"github.com/tdewolff/minify/xml"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/extension"
)

// Minify chainable context.
type Minify interface {
	goldsmith.Plugin
	goldsmith.Initializer
	goldsmith.Processor
}

// New creates a new instance of the Minify plugin
func New() Minify {
	return new(minify)
}

type minify struct {
}

func (*minify) Name() string {
	return "minify"
}

func (*minify) Initialize(ctx *goldsmith.Context) ([]goldsmith.Filter, error) {
	return []goldsmith.Filter{extension.New(".css", ".html", ".htm", ".js", ".svg", ".json", ".xml")}, nil
}

func (*minify) Process(ctx *goldsmith.Context, f *goldsmith.File) error {
	var (
		buff bytes.Buffer
		err  error
	)

	switch m := min.New(); filepath.Ext(f.Path()) {
	case ".css":
		err = css.Minify(m, &buff, f, nil)
	case ".html", ".htm":
		err = html.Minify(m, &buff, f, nil)
	case ".js":
		err = js.Minify(m, &buff, f, nil)
	case ".json":
		err = json.Minify(m, &buff, f, nil)
	case ".svg":
		err = svg.Minify(m, &buff, f, nil)
	case ".xml":
		err = xml.Minify(m, &buff, f, nil)
	}

	if err != nil {
		return err
	}

	nf := goldsmith.NewFileFromData(f.Path(), buff.Bytes())
	nf.InheritValues(f)
	ctx.DispatchFile(nf, false)

	return nil
}
