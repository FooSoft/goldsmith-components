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

func (*minify) Initialize(context *goldsmith.Context) ([]goldsmith.Filter, error) {
	return []goldsmith.Filter{extension.New(".css", ".html", ".htm", ".js", ".svg", ".json", ".xml")}, nil
}

func (*minify) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	if outputFile := context.RetrieveCachedFile(inputFile.Path(), inputFile); outputFile != nil {
		context.DispatchFile(outputFile)
		return nil
	}

	var (
		buff bytes.Buffer
		err  error
	)

	switch m := min.New(); filepath.Ext(inputFile.Path()) {
	case ".css":
		err = css.Minify(m, &buff, inputFile, nil)
	case ".html", ".htm":
		err = html.Minify(m, &buff, inputFile, nil)
	case ".js":
		err = js.Minify(m, &buff, inputFile, nil)
	case ".json":
		err = json.Minify(m, &buff, inputFile, nil)
	case ".svg":
		err = svg.Minify(m, &buff, inputFile, nil)
	case ".xml":
		err = xml.Minify(m, &buff, inputFile, nil)
	}

	if err != nil {
		return err
	}

	outputFile := goldsmith.NewFileFromData(inputFile.Path(), buff.Bytes())
	outputFile.InheritValues(inputFile)
	context.DispatchAndCacheFile(outputFile)

	return nil
}
