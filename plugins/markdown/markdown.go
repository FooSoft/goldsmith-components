// Package markdown renders Markdown documents to HTML with the "blackfriday"
// processor. You can specify which Markdown extensions and HTML features to
// use by directly passing the blackfriday flags to this plugin. Note that
// unlike other static site generators, Markdown processing does not
// automatically parse frontmatter; you will need to use the "frontmatter"
// plugin to extract any metadata which may be present in your source content.
package markdown

import (
	"bytes"
	"path"
	"strings"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/wildcard"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

// Markdown chainable context.
type Markdown struct {
}

// New creates a new instance of the Markdown plugin.
func New() *Markdown {
	return new(Markdown)
}

func (*Markdown) Name() string {
	return "markdown"
}

func (plugin *Markdown) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	return wildcard.New("**/*.md", "**/*.markdown"), nil
}

func (plugin *Markdown) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	outputPath := strings.TrimSuffix(inputFile.Path(), path.Ext(inputFile.Path())) + ".html"
	if outputFile := context.RetrieveCachedFile(outputPath, inputFile); outputFile != nil {
		outputFile.Meta = inputFile.Meta
		context.DispatchFile(outputFile)
		return nil
	}

	var dataIn bytes.Buffer
	if _, err := dataIn.ReadFrom(inputFile); err != nil {
		return err
	}

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	)

	var dataOut bytes.Buffer
	if err := md.Convert(dataIn.Bytes(), &dataOut); err != nil {
		return err
	}

	outputFile := context.CreateFileFromData(outputPath, dataOut.Bytes())
	outputFile.Meta = inputFile.Meta
	context.DispatchAndCacheFile(outputFile, inputFile)
	return nil
}
