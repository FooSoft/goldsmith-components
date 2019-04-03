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
	"github.com/FooSoft/goldsmith-components/filters/wildcard"
)

// Minify chainable context.
type Minify struct{}

// New creates a new instance of the Minify plugin
func New() *Minify {
	return new(Minify)
}

func (*Minify) Name() string {
	return "minify"
}

func (*Minify) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	return wildcard.New("**/*.css", "**/*.html", "**/*.htm", "**/*.js", "**/*.svg", "**/*.json", "**/*.xml"), nil
}

func (*Minify) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	if outputFile := context.RetrieveCachedFile(inputFile.Path(), inputFile); outputFile != nil {
		outputFile.Meta = inputFile.Meta
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

	outputFile := context.CreateFileFromData(inputFile.Path(), buff.Bytes())
	outputFile.Meta = inputFile.Meta
	context.DispatchAndCacheFile(outputFile, inputFile)
	return nil
}
