// Package markdown renders Markdown documents to HTML with the "goldmark"
// processor. Note that unlike other static site generators, Markdown processing
// does not automatically parse frontmatter; you will need to use the "frontmatter"
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
	"github.com/yuin/goldmark/renderer/html"
)

// Markdown chainable context.
type Markdown struct {
	md goldmark.Markdown
}

// New creates a new instance of the Markdown plugin.
func New() *Markdown {
	return NewWithGoldmark(
		goldmark.New(
			goldmark.WithExtensions(extension.GFM, extension.Typographer),
			goldmark.WithParserOptions(parser.WithAutoHeadingID()),
			goldmark.WithRendererOptions(html.WithUnsafe()),
		),
	)
}

// New creates a new instance of the Markdown plugin with user-provided goldmark instance.
func NewWithGoldmark(md goldmark.Markdown) *Markdown {
	return &Markdown{md}
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

	var dataOut bytes.Buffer
	if err := plugin.md.Convert(dataIn.Bytes(), &dataOut); err != nil {
		return err
	}

	outputFile := context.CreateFileFromData(outputPath, dataOut.Bytes())
	outputFile.Meta = inputFile.Meta
	context.DispatchAndCacheFile(outputFile, inputFile)
	return nil
}
